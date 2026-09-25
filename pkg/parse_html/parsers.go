package parse_html

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/net/html/charset"
)

var (
	skipExt = map[string]struct{}{
		".pdf": {}, ".doc": {}, ".docx": {}, ".rtf": {}, ".odt": {},
		".xls": {}, ".xlsx": {}, ".ods": {}, ".ppt": {}, ".pptx": {},
		".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {},
		".svg": {}, ".ico": {}, ".bmp": {}, ".tiff": {},
		".mp3": {}, ".mp4": {}, ".avi": {}, ".mov": {}, ".mkv": {}, ".webm": {},
		".zip": {}, ".rar": {}, ".7z": {}, ".tar": {}, ".gz": {}, ".bz2": {},
		".exe": {}, ".dmg": {}, ".apk": {},
		".css": {}, ".js": {}, ".json": {}, ".xml": {}, ".rss": {}, ".atom": {},
		".woff": {}, ".woff2": {}, ".ttf": {}, ".eot": {},
	}

	ExtError    = errors.New("unsupported file extension")
	SchemeError = errors.New("unsupported url scheme")
)

func Parse(resp *http.Response) (*html.Node, error) {
	defer resp.Body.Close()
	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("charset reader: %w", err)
	}

	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func ParseLinks(doc *html.Node) ([]string, error) {
	links := make([]string, 0)

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.A && n.FirstChild != nil {
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
		if n.Type == html.ElementNode && n.DataAtom == atom.Title && n.FirstChild != nil {
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

	ext := strings.ToLower(path.Ext(absolute.Path))
	if _, skip := skipExt[ext]; skip {
		return "", ExtError
	}

	if absolute.Scheme != "http" && absolute.Scheme != "https" {
		return "", SchemeError
	}

	return absolute.String(), nil
}
