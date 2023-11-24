package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

func Init(c *Config) zerolog.Logger {
	zerolog.TimeFieldFormat = lo.Ternary(c.Timestamp.Enabled, c.Timestamp.Format, "")

	switch c.Level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	var logger zerolog.Logger
	switch c.Target {
	case "stdout":
		switch c.Format {
		case "json":
			if c.Caller {
				return zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
			}
			return zerolog.New(os.Stdout).With().Timestamp().Logger()
		case "console":
			if c.Caller {
				return zerolog.New(os.Stdout).With().Timestamp().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})
			}
			return zerolog.New(os.Stdout).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})
		}
	default:
		switch c.Format {
		case "json":
			if c.Caller {
				return zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
			}
			return zerolog.New(os.Stderr).With().Timestamp().Logger()
		case "console":
			if c.Caller {
				return zerolog.New(os.Stderr).With().Timestamp().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
			}
			return zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
		}
	}
	return logger
}
