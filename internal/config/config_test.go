package config

import (
	"flag"
	"testing"
	"time"

	"github.com/urfave/cli"
)

func TestNew_MapsFlagsToConfig(t *testing.T) {
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String("urls", "https://a.com,https://b.com", "")
	set.Uint("depth", 3, "")
	set.Duration("timeout", 2*time.Minute, "")
	set.Duration("request-timeout", 10*time.Second, "")
	set.String("output", "out.json", "")
	set.String("log", "crawler.log", "")
	set.String("log-level", "debug", "")
	set.Uint("go", 5, "")

	ctx := cli.NewContext(nil, set, nil)

	cfg, err := New(*ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantUrls := Urls{"https://a.com", "https://b.com"}
	if len(cfg.Urls) != len(wantUrls) {
		t.Fatalf("urls length mismatch: got %v, want %v", cfg.Urls, wantUrls)
	}
	for i := range wantUrls {
		if cfg.Urls[i] != wantUrls[i] {
			t.Errorf("url[%d] = %q, want %q", i, cfg.Urls[i], wantUrls[i])
		}
	}

	if cfg.Depth != 3 {
		t.Errorf("Depth = %d, want 3", cfg.Depth)
	}
	if cfg.Timeout != 2*time.Minute {
		t.Errorf("Timeout = %v, want 2m", cfg.Timeout)
	}
	if cfg.ReqTimeout != 10*time.Second {
		t.Errorf("ReqTimeout = %v, want 10s", cfg.ReqTimeout)
	}
	if cfg.Output != "out.json" {
		t.Errorf("Output = %q, want %q", cfg.Output, "out.json")
	}
	if cfg.Goroutines != 5 {
		t.Errorf("Goroutines = %d, want 5", cfg.Goroutines)
	}
	if cfg.Logger.Level != "debug" {
		t.Errorf("Logger.Level = %q, want %q", cfg.Logger.Level, "debug")
	}
	if cfg.Logger.OutputFile != "crawler.log" {
		t.Errorf("Logger.OutputFile = %q, want %q", cfg.Logger.OutputFile, "crawler.log")
	}
}

func TestNew_SingleURLNoComma(t *testing.T) {
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String("urls", "https://only.com", "")
	set.Uint("depth", 1, "")
	set.Duration("timeout", time.Minute, "")
	set.Duration("request-timeout", time.Second, "")
	set.String("output", "o.json", "")
	set.String("log", "", "")
	set.String("log-level", "info", "")
	set.Uint("go", 1, "")

	ctx := cli.NewContext(nil, set, nil)

	cfg, err := New(*ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Urls) != 1 || cfg.Urls[0] != "https://only.com" {
		t.Errorf("got %v, want a single-element slice with https://only.com", cfg.Urls)
	}
}
