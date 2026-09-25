package concurent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"github.com/hlebecz/crawler-cli/internal/client"
	"github.com/hlebecz/crawler-cli/internal/config"
	"github.com/hlebecz/crawler-cli/internal/crawler"
	"github.com/hlebecz/crawler-cli/internal/model"
	"github.com/hlebecz/crawler-cli/pkg/parse_html"
)

type Output interface {
	Output(n model.Node) error
}

type Cache interface {
	ShouldVisit(url string) bool
}

type Crawler struct {
	crawler.BaseCrawler
}

func New(config config.Config, client client.Client, output crawler.Output, cache crawler.Cache) *Crawler {
	return &Crawler{
		crawler.BaseCrawler{
			Config: config,
			Client: client,
			Output: output,
			Cache:  cache,
		},
	}
}

func (c *Crawler) CrawlAll(ctx context.Context) {
	urls := c.Config.Urls
	if len(urls) == 0 {
		return
	}

	startNodes := make([]*model.Node, 0, len(urls))
	unvisited := make(chan *model.Node, 10_000)
	visited := make(chan *model.Node, 1_000)

	var (
		wg        sync.WaitGroup
		pending   atomic.Int64
		closeOnce sync.Once
	)
	closeIn := func() { closeOnce.Do(func() { close(unvisited) }) }

	for _, u := range urls {
		node := model.NewNode(u, "", 1)
		startNodes = append(startNodes, &node)
		pending.Add(1)
		unvisited <- &node
	}

	for range c.Config.Goroutines {
		wg.Go(func() {
			c.worker(ctx, unvisited, visited, &pending, closeIn, startNodes)
		})
	}

	go func() {
		wg.Wait()
		close(visited)
	}()

	var buf []*model.Node
FOR:
	for {
		var sendCh chan *model.Node
		var next *model.Node
		if len(buf) > 0 {
			sendCh = unvisited
			next = buf[0]
		}

		select {
		case n, ok := <-visited:
			if !ok {
				break FOR
			}
			for _, link := range n.Links {
				select {
				case unvisited <- link:

				default:
					buf = append(buf, link)
				}
			}
		case sendCh <- next:
			buf = buf[1:]
			if len(buf) == 0 {
				buf = nil
			}
		case <-ctx.Done():
			closeIn()
			for range visited {
			}
			break FOR
		}
	}

	if err := c.Output.Output(startNodes); err != nil {
		log.Err(err).Msg("failed to output a node")
	}
}

func (c *Crawler) worker(
	ctx context.Context,
	unvisited <-chan *model.Node,
	visited chan<- *model.Node,
	pending *atomic.Int64,
	closeIn func(),
	startNodes []*model.Node,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case n, ok := <-unvisited:
			if !ok {
				return
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error().Interface("panic", r).Str("url", n.Resource).Msg("worker panic")
					}
				}()
				c.crawl(ctx, n, visited, pending, startNodes)
			}()
			if pending.Add(-1) == 0 {
				closeIn()
			}
		}
	}
}

func (c *Crawler) crawl(
	ctx context.Context,
	n *model.Node,
	out chan<- *model.Node,
	pending *atomic.Int64,
	startNodes []*model.Node,
) {
	resp, err := c.Client.Get(ctx, n.Resource)
	if err != nil {
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			log.Warn().Err(err).Str("url", n.Resource).Msg("http get failed")
		}
		return
	}

	doc, err := parse_html.Parse(resp)
	if err != nil {
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			log.Warn().Err(err).Str("url", n.Resource).Msg("parse failed")
		}
		return
	}

	n.Title, err = parse_html.ParseTitle(doc)
	if err != nil {
		log.Warn().Err(err).Str("url", n.Resource).Msg("parse title failed")
	}

	if n.Depth < c.Config.Depth {
		links, err := parse_html.ParseLinks(doc)
		if err != nil {
			log.Warn().Err(err).Str("url", n.Resource).Msg("parse links failed")
		} else if childs, err := c.CreateEmptyNodes(links, n.Depth+1, n.Resource, startNodes); err != nil {
			log.Warn().Err(err).Str("url", n.Resource).Msg("create nodes failed")
		} else {
			for _, ch := range childs {
				n.Links = append(n.Links, &ch)
			}
			pending.Add(int64(len(n.Links)))
		}
	}

	select {
	case out <- n:
		//log.Debug().Str("url", n.Resource).Msg("finished crawling node")
	case <-ctx.Done():
		if len(n.Links) > 0 {
			pending.Add(-int64(len(n.Links)))
		}
	}
}
