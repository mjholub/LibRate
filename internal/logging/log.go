package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func Init() zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC822Z
	switch os.Getenv("LOG_LEVEL") {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel) // pretty, for dev env
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel) // compact, for prod env
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if os.Getenv("LOG_TARGET") == "stdout" {
		if os.Getenv("LOG_FORMAT") == "json" {
			return zerolog.New(os.Stdout).With().Timestamp().
				Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})
		}
		return zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	if os.Getenv("LOG_FORMAT") == "json" {
		return zerolog.New(os.Stderr).With().Timestamp().
			Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	return zerolog.New(os.Stderr).With().Timestamp().Logger()
}