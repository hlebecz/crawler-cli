package internal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/hlebecz/crawler-cli/internal/adapter/cache/simplecache"
	"github.com/hlebecz/crawler-cli/internal/adapter/output/fileoutput"
	"github.com/hlebecz/crawler-cli/internal/client"
	"github.com/hlebecz/crawler-cli/internal/config"
	"github.com/hlebecz/crawler-cli/internal/crawler/concurent"
	"github.com/hlebecz/crawler-cli/internal/crawler/recursive"
)

func Run(ctx context.Context, c config.Config) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	file, err := os.OpenFile(c.Output, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create output file")
	}
	out := fileoutput.New(file)

	client := client.New(10, c.ReqTimeout)

	cache := simplecache.New()

	done := make(chan struct{})

	go func() {
		defer close(done)
		if c.Goroutines == 1 {
			crawler := recursive.New(c, client, out, cache)
			crawler.CrawlAll(ctx)
		} else {
			crawler := concurent.New(c, client, out, cache)
			crawler.CrawlAll(ctx)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	select {
	case <-sig:
		log.Info().Msg("received signal, shutting down gracefully")
		cancel()
		<-done
	case <-done:
		log.Info().Msg("crawler is done")
	}

	err = out.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to close output file")
	}
}
