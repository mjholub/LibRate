package cfg

import (
	"fmt"

	"codeberg.org/mjh/LibRate/cfg/parser"
	"codeberg.org/mjh/LibRate/internal/clitools"
	"codeberg.org/mjh/LibRate/internal/logging"

	config "github.com/gookit/config/v2"
	"github.com/imdario/mergo"
	"github.com/samber/mo"
)

// nolint:gochecknoglobals
var log = logging.Init(&logging.Config{
	Level:  "info",
	Target: "stdout",
	Format: "json",
})

func LoadConfig() mo.Result[*Config] {
	return mo.Try(func() (conf *Config, err error) {
		// first, look for an existing config file
		if confLoc := lookForExisting(tryLocations()); confLoc != "" {
			log.Info().Msgf("found config at %s", confLoc)
			conf, err = parseRaw(confLoc)
			if err != nil {
				return conf, fmt.Errorf("failed to parse config: %w", err)
			}
			if err = mergo.Merge(&conf, &Config{}); err != nil {
				return conf, fmt.Errorf("failed to merge config structs: %w", err)
			}
			return conf, nil
		}
		// if not found, fall back to defaults
		defaultConfig, err := parseRaw("example_config.yml")
		if err != nil {
			return conf, fmt.Errorf("failed to parse default config: %w", err)
		}
		log.Info().Msgf("using default config: %v", defaultConfig)

		if err := mergo.Merge(&conf, defaultConfig); err != nil {
			return conf, fmt.Errorf("failed to merge config structs: %w", err)
		}

		return conf, nil
	})
}

func parseRaw(configLocation string) (conf *Config, err error) {
	configRaw, err := parser.Parse(configLocation)
	if err != nil {
		return nil,
			fmt.Errorf("failed to parse config: %w", err)
	}
	log.Info().Msgf("got config: %v", configRaw)

	configStr := createKVPairs(configRaw)
	_ = config.MapStruct(configStr, conf)

	return conf, nil
}

func tryGettingConfig(tryPaths []string) (string, error) {
	defaultConfigPath, err := getDefaultConfigPath()
	if err != nil {
		return "", err
	}
	customPath, err := clitools.AskPath("config", defaultConfigPath, tryPaths)
	if err != nil {
		return defaultConfigPath, err
	}

	return customPath, nil
}
