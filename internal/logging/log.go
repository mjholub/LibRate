package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// FIXME: add function to load logger config in the config package
func Init(c *Config) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC822Z
	switch c.Level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel) // pretty, for dev env
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel) // compact, for prod env
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if c.Target == "stdout" {
		if c.Format == "json" {
			return zerolog.New(os.Stdout).With().Timestamp().
				Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})
		}
		return zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	if c.Format == "json" {
		return zerolog.New(os.Stderr).With().Timestamp().
			Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	return zerolog.New(os.Stderr).With().Timestamp().Logger()
}
