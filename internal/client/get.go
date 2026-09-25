package client

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/rs/zerolog/log"
)

func (c Client) Get(ctx context.Context, url string) (*http.Response, error) {
	select {
	case c.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	defer func() { <-c.sem }()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,application/xml;q=0.8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	log.Debug().Str("url", url).Int("code", resp.StatusCode).Msg("got response")
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	mt, _, _ := mime.ParseMediaType(ct)
	switch mt {
	case "text/html", "application/xhtml+xml", "application/xml", "text/xml":
	default:
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		return nil, ErrNotHTML
	}

	return resp, nil
}
