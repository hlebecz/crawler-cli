package parse_html

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestResolveURL_RelativeLink(t *testing.T) {
	got, err := ResolveURL("https://example.com/dir/page.html", "sub/page2.html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://example.com/dir/sub/page2.html"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveURL_SkipsBlockedExtensions(t *testing.T) {
	_, err := ResolveURL("https://example.com/", "file.pdf")
	if !errors.Is(err, ExtError) {
		t.Errorf("expected ExtError, got %v", err)
	}
}

func TestResolveURL_RejectsNonHTTPScheme(t *testing.T) {
	_, err := ResolveURL("https://example.com/", "mailto:test@example.com")
	if !errors.Is(err, SchemeError) {
		t.Errorf("expected SchemeError, got %v", err)
	}
}

func TestResolveURL_KeepsPlainHTMLLink(t *testing.T) {
	got, err := ResolveURL("https://example.com/", "/about")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com/about" {
		t.Errorf("got %q, want %q", got, "https://example.com/about")
	}
}

func TestParseTitle(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`<html><head><title>Hello World</title></head><body></body></html>`))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	title, err := ParseTitle(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != "Hello World" {
		t.Errorf("got %q, want %q", title, "Hello World")
	}
}

func TestParseTitle_MissingFallsBackToDefault(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`<html><body>no title here</body></html>`))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	title, err := ParseTitle(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != "No title" {
		t.Errorf("got %q, want fallback %q", title, "No title")
	}
}

func TestParseLinks(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`
		<html><body>
			<a href="/a">A</a>
			<a href="/b">B</a>
			<a>no href, should be ignored</a>
		</body></html>
	`))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	links, err := ParseLinks(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("got %d links, want 2: %v", len(links), links)
	}
	if links[0] != "/a" || links[1] != "/b" {
		t.Errorf("unexpected links: %v", links)
	}
}
