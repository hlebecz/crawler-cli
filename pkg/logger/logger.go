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

type LevelWriter struct {
	io.Writer
	Level zerolog.Level
}

func (lw LevelWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	if l >= lw.Level {
		return lw.Writer.Write(p)
	}
	return len(p), nil
}

func Init(c Config) {
	zerolog.TimeFieldFormat = time.RFC3339

	level, err := zerolog.ParseLevel(c.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	//zerolog.SetGlobalLevel(level)

	writers := make([]io.Writer, 0, 2)

	if c.PrettyConsole {
		writers = append(writers, LevelWriter{
			&zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"},
			level,
		})
	} else {
		writers = append(writers, LevelWriter{
			zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(level),
			level,
		})
	}

	if c.OutputFile != "" {
		f, err := os.OpenFile(c.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to open log file")
		} else {
			wr := zerolog.New(f).With().Timestamp().Logger()
			writers = append(writers, LevelWriter{wr, level - 1})
		}
	}
	log.Logger = log.Output(zerolog.MultiLevelWriter(writers...))

	log.Info().Msg("Logger initialized")
}
