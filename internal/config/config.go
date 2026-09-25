package config

import (
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/hlebecz/crawler-cli/pkg/logger"
)

type Urls []string

type Config struct {
	Urls       Urls
	Depth      uint
	Timeout    time.Duration
	ReqTimeout time.Duration
	Output     string
	Logger     logger.Config
	Goroutines uint
}

func New(c cli.Context) (Config, error) {
	return Config{
		Urls:       strings.Split(c.String("urls"), ","),
		Depth:      c.Uint("depth"),
		Timeout:    c.Duration("timeout"),
		ReqTimeout: c.Duration("request-timeout"),
		Output:     c.String("output"),
		Logger: logger.Config{
			Level:         c.String("log-level"),
			PrettyConsole: true,
			OutputFile:    c.String("log"),
		},
		Goroutines: c.Uint("go"),
	}, nil
}
