package crawler

import (
	"errors"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/hlebecz/crawler-cli/internal/client"
	"github.com/hlebecz/crawler-cli/internal/config"
	"github.com/hlebecz/crawler-cli/internal/model"
	"github.com/hlebecz/crawler-cli/pkg/parse_html"
)

type Output interface {
	Output(n []*model.Node) error
}

type Cache interface {
	ShouldVisit(url string) bool
}

type BaseCrawler struct {
	Config config.Config
	Client client.Client
	Output Output
	Cache  Cache
}

func (c *BaseCrawler) CreateEmptyNodes(links []string, depth uint, base string, startNodes []*model.Node) ([]model.Node, error) {
	nodes := make([]model.Node, 0, len(links))

	for _, l := range links {
		url, err := parse_html.ResolveURL(base, l)
		if err != nil {
			if errors.Is(err, parse_html.ExtError) || errors.Is(err, parse_html.SchemeError) {
				log.Debug().Err(err).Str("url", url).Msg("skipping url")
			} else {
				log.Warn().Err(err).Str("url", l).Msg("failed to resolve url")
			}
		}

		if !c.Cache.ShouldVisit(url) || !c.checkBelonging(url, startNodes) {
			continue
		}
		nodes = append(nodes, model.NewNode(url, "", depth))
	}

	return nodes, nil
}

func (c *BaseCrawler) checkBelonging(url string, startNodes []*model.Node) bool {
	for _, n := range startNodes {
		if strings.Contains(url, n.Resource) {
			return true
		}
	}
	return false
}
