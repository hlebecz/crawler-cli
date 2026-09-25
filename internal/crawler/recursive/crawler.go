package recursive

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/hlebecz/crawler-cli/internal/client"
	"github.com/hlebecz/crawler-cli/internal/config"
	"github.com/hlebecz/crawler-cli/internal/crawler"
	"github.com/hlebecz/crawler-cli/internal/model"
	"github.com/hlebecz/crawler-cli/pkg/parse_html"
)

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

	startNodes := make([]*model.Node, 0, len(urls))

	for _, u := range urls {
		node := model.NewNode(u, "", 1)
		startNodes = append(startNodes, &node)
		err := c.crawl(ctx, &node, startNodes)
		if err != nil {
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				log.Err(err).Msg("failed to crawl")
				break
			}
			continue
		}
	}

	err := c.Output.Output(startNodes)
	if err != nil {
		log.Err(err).Msg("failed to output a node")
	}
}

func (c *Crawler) crawl(ctx context.Context, n *model.Node, startNodes []*model.Node) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	resp, err := c.Client.Get(ctx, n.Resource)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}

	doc, err := parse_html.Parse(resp)
	if err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	n.Title, err = parse_html.ParseTitle(doc)
	if err != nil {
		log.Warn().Err(err).Str("url", n.Resource).Msg("unable to parse title")
	}

	if n.Depth > c.Config.Depth {
		//log.Debug().Msgf("Skipping crawl of depth %d", n.Depth)
		return nil
	}

	links, err := parse_html.ParseLinks(doc)
	if err != nil {
		log.Warn().Err(err).Str("url", n.Resource).Msg("unable to parse links")
	}

	childs, err := c.CreateEmptyNodes(links, n.Depth+1, n.Resource, startNodes)
	if err != nil {
		log.Warn().Err(err).Str("url", n.Resource).Msg("unable to create new nodes")
	}

	for _, ch := range childs {
		err = c.crawl(ctx, &ch, startNodes)
		if err != nil {
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				log.Warn().Err(err).Str("url", n.Resource).Msg("unable to crawl")
			}
		}
		n.Links = append(n.Links, &ch)
	}

	//log.Debug().Str("url", n.Resource).Msg("finished crawling node")

	return nil
}
