package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Level         string
	PrettyConsole bool
	OutputFile    string
}

func Init(c Config) {
	zerolog.TimeFieldFormat = time.RFC3339

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	level, err := zerolog.ParseLevel(c.Level)
	if err != nil {
		zerolog.SetGlobalLevel(level)
	}

	writers := make([]io.Writer, 0, 2)

	if c.PrettyConsole {
		writers = append(writers, &zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	} else {
		writers = append(writers, log.With().Caller().Logger())
	}

	f, err := os.OpenFile(c.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to open log file")
	} else {
		writers = append(writers, f)
	}
	log.Logger = log.Output(zerolog.MultiLevelWriter(writers...))

	log.Info().Msg("Logger initialized")
}
