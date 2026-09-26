package crawler

import (
	"testing"

	"github.com/hlebecz/crawler-cli/internal/model"
)

type alwaysVisitCache struct{}

func (alwaysVisitCache) ShouldVisit(url string) bool { return true }

type neverVisitCache struct{}

func (neverVisitCache) ShouldVisit(url string) bool { return false }

func TestCreateEmptyNodes_KeepsInDomainLinks(t *testing.T) {
	c := &BaseCrawler{Cache: alwaysVisitCache{}}

	start := model.NewNode("https://example.com", "", 1)
	startNodes := []*model.Node{&start}

	links := []string{"/about", "/contact"}

	nodes, err := c.CreateEmptyNodes(links, 2, "https://example.com", startNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("got %d nodes, want 2: %+v", len(nodes), nodes)
	}
}

func TestCreateEmptyNodes_DropsUnsupportedExtensions(t *testing.T) {
	c := &BaseCrawler{Cache: alwaysVisitCache{}}

	start := model.NewNode("https://example.com", "", 1)
	startNodes := []*model.Node{&start}

	links := []string{"/doc.pdf", "/page.html"}

	nodes, err := c.CreateEmptyNodes(links, 2, "https://example.com", startNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1 (pdf should be filtered out): %+v", len(nodes), nodes)
	}
	if nodes[0].Resource != "https://example.com/page.html" {
		t.Errorf("got %q, want %q", nodes[0].Resource, "https://example.com/page.html")
	}
}

func TestCreateEmptyNodes_SkipsAlreadyVisitedURLs(t *testing.T) {
	c := &BaseCrawler{Cache: neverVisitCache{}}

	start := model.NewNode("https://example.com", "", 1)
	startNodes := []*model.Node{&start}

	nodes, err := c.CreateEmptyNodes([]string{"/page"}, 2, "https://example.com", startNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes for an already-visited URL, got %d", len(nodes))
	}
}

func TestCreateEmptyNodes_SetsCorrectDepth(t *testing.T) {
	c := &BaseCrawler{Cache: alwaysVisitCache{}}

	start := model.NewNode("https://example.com", "", 1)
	startNodes := []*model.Node{&start}

	nodes, err := c.CreateEmptyNodes([]string{"/page"}, 3, "https://example.com", startNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	if nodes[0].Depth != 3 {
		t.Errorf("Depth = %d, want 3", nodes[0].Depth)
	}
}
