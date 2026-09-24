package parse_html

import (
	"fmt"
	"net/url"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ParseLinks(doc *html.Node) ([]string, error) {
	links := make([]string, 0)

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.A {
			for _, a := range n.Attr {
				if a.Key == "href" {
					links = append(links, a.Val)
					break
				}
			}
		}
	}

	return links, nil
}

func ParseTitle(doc *html.Node) (string, error) {
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.Title {
			return n.FirstChild.Data, nil
		}
	}
	return "No title", nil
}

func ResolveURL(baseURL, href string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	ref, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	absolute := base.ResolveReference(ref)

	if absolute.Scheme != "http" && absolute.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme: %s", absolute.Scheme)
	}

	return absolute.String(), nil
}
