package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

// FIXME: add function to load logger config in the config package
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

	logger := lo.TernaryF(c.Target == "stdout",
		func() interface{} {
			return lo.TernaryF(c.Format == "json",
				func() interface{} {
					return zerolog.New(os.Stdout).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})
				},
				func() interface{} {
					return zerolog.New(os.Stdout).With().Timestamp().Logger()
				}).(zerolog.Logger)
		},
		func() interface{} {
			return lo.TernaryF(c.Format == "json",
				func() interface{} {
					return zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
				},
				func() interface{} {
					return zerolog.New(os.Stderr).With().Timestamp().Logger()
				}).(zerolog.Logger)
		}).(zerolog.Logger)

	if c.Caller {
		logger = logger.With().Caller().Logger()
	}

	return logger
}
