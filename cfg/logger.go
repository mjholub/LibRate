package cfg

import (
	"fmt"

	config "github.com/gookit/config/v2"
	"github.com/samber/mo"

	"codeberg.org/mjh/LibRate/cfg/parser"
	"codeberg.org/mjh/LibRate/internal/logging"
)

// TODO: refactor so that parser parses each section of the config separately
func LoadLoggerConfig() mo.Result[logging.Config] {
	return mo.Try(func() (logging.Config, error) {
		loc := lookForExisting(tryLocations())
		configRaw, err := parser.Parse(loc)
		if err != nil {
			return logging.Config{}, fmt.Errorf("failed to parse logger config: %w", err)
		}
		log.Info().Msgf("got logger config: %v", configRaw)

		// preallocate default config
		conf := logging.Config{
			Level:  "info",
			Target: "stdout",
			Format: "json",
		}

		configStr := createKVPairs(configRaw)
		_ = config.MapStruct(configStr, conf) // WARN: unsure if this is correct

		return conf, nil
	})
}
