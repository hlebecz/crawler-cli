package config

import (
	"time"

	"github.com/urfave/cli"

	"github.com/hlebecz/crawler-cli/pkg/logger"
)

type Urls []string

type Config struct {
	Urls           Urls
	Depth          int
	Timeout        time.Duration
	RequestTimeout time.Duration
	Output         string
	Logger         logger.Config
}

func New(c cli.Context) (Config, error) {
	return Config{
		Urls:           c.StringSlice("urls"),
		Depth:          c.Int("depth"),
		Timeout:        c.Duration("timeout"),
		RequestTimeout: c.Duration("request-timeout"),
		Output:         c.String("output"),
		Logger: logger.Config{
			Level:         c.String("log-level"),
			PrettyConsole: true,
			OutputFile:    c.String("log"),
		},
	}, nil
}
