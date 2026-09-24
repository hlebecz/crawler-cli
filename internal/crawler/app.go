package crawler

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/hlebecz/crawler-cli/internal/cache"
	"github.com/hlebecz/crawler-cli/internal/client"
	"github.com/hlebecz/crawler-cli/internal/config"
	"github.com/hlebecz/crawler-cli/internal/crawler/recursive"
	"github.com/hlebecz/crawler-cli/internal/output"
)

func Run(ctx context.Context, c config.Config) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	file, err := os.OpenFile("output.json", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create output file")
	}
	out := output.NewFileOutput(file)

	client := client.NewClient()

	cache := cache.NewSimpleCache()

	crawler := recursive.NewCrawler(c, client, out, cache)

	crawler.CrawlAll(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig

	err = out.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to close output file")
	}
}
