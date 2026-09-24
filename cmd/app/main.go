package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"

	"github.com/hlebecz/crawler-cli/internal/config"
	"github.com/hlebecz/crawler-cli/internal/crawler"
	"github.com/hlebecz/crawler-cli/pkg/logger"
)

func main() {
	ctx := context.Background()
	app := cli.NewApp()
	app.Name = "crawler"
	app.Usage = "crawler is a url parsing tool"
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{Name: "urls", Required: true, Usage: "url to crawl"},
		cli.IntFlag{Name: "depth", Value: 2, Usage: "depth"},
		cli.DurationFlag{Name: "timeout", Value: time.Minute * 2, Usage: "timeout"},
		cli.DurationFlag{Name: "request-timeout", Value: time.Second * 10, Usage: "timeout"},
		cli.StringFlag{Name: "output", Value: "output.json", Usage: "output file name"},
		cli.StringFlag{Name: "log", Value: "crawler.log", Usage: "log file path"},
		cli.StringFlag{Name: "error", Value: "error", Usage: "log level"},
	}
	app.Action = func(c *cli.Context) error {
		cfg, err := config.New(*c)
		if err != nil {
			return err
		}
		logger.Init(cfg.Logger)
		crawler.Run(ctx, cfg)
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run app")
	}
}
