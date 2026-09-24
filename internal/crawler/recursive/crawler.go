package recursive

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"

	"github.com/hlebecz/crawler-cli/internal/client"
	"github.com/hlebecz/crawler-cli/internal/config"
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
	Config config.Config
	Client *client.Client
	Output Output
	Cache  Cache
}

func NewCrawler(config config.Config, client *client.Client, output Output, cache Cache) *Crawler {
	return &Crawler{
		Config: config,
		Client: client,
		Output: output,
		Cache:  cache,
	}
}

func (c *Crawler) crawl(ctx context.Context, n *model.Node) error {
	resp, err := c.Client.Get(n.Resource)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	n.Title, err = parse_html.ParseTitle(doc)
	if err != nil {
		log.Err(err).Str("url", n.Resource).Msg("unable to parse title")
	}

	if n.Depth > c.Config.Depth {
		log.Warn().Msgf("Skipping crawl of depth %d", n.Depth)

		return nil
	}

	links, err := parse_html.ParseLinks(doc)
	if err != nil {
		log.Err(err).Str("url", n.Resource).Msg("unable to parse links")
	}

	childs, err := c.createEmptyNodes(links, n.Depth+1, n.Resource)
	if err != nil {
		log.Err(err).Str("url", n.Resource).Msg("unable to create new nodes")
	}

	for _, ch := range childs {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = c.crawl(ctx, &ch)
		if err != nil {
			log.Err(err).Str("url", n.Resource).Msg("unable to crawl")
		}
		n.Links = append(n.Links, ch)
	}

	log.Info().Msgf("finished crawling node: %s", n.Resource)

	return nil
}

func (c *Crawler) CrawlAll(ctx context.Context) {
	urls := c.Config.Urls

	startNodes := make([]*model.Node, 0, len(urls))

	for _, u := range urls {
		node := model.NewNode(u, "", 1)
		startNodes = append(startNodes, &node)
		err := c.crawl(ctx, &node)
		if err != nil {
			log.Err(err).Msg("failed to crawl")
		}
	}

	for _, node := range startNodes {
		err := c.Output.Output(*node)
		if err != nil {
			log.Err(err).Msg("failed to output a node")
		}
	}
}

func (c Crawler) createEmptyNodes(links []string, depth int, base string) ([]model.Node, error) {
	nodes := make([]model.Node, 0, len(links))

	for _, l := range links {
		url, err := parse_html.ResolveURL(base, l)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve url: %w", err)
		}
		if !c.Cache.ShouldVisit(url) {
			continue
		}
		nodes = append(nodes, model.NewNode(url, "", depth))
	}

	return nodes, nil
}
