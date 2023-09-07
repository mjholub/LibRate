package cfg

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"codeberg.org/mjh/LibRate/cfg/parser"
	"codeberg.org/mjh/LibRate/internal/clitools"
	"codeberg.org/mjh/LibRate/internal/logging"

	config "github.com/gookit/config/v2"
	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
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
		conf = &Config{}
		if confLoc, err := lookForExisting(tryLocations()); err == nil && confLoc != "" {
			log.Info().Msgf("found config at %s", confLoc)
			loadedConfig, err := parseRaw(confLoc)
			if err != nil {
				return conf, fmt.Errorf("failed to parse config: %w", err)
			}
			// FIXME: the db config is not getting properly marshalled into the struct
			log.Debug().
				Msgf("DB config from the conf variable before merge: %+v",
					conf.DBConfig)
			log.Debug().
				Msgf("DB config from the loadedConfig variable before merge: %+v",
					loadedConfig.DBConfig)
			if err = mergo.Merge(conf, loadedConfig); err != nil {
				return conf, fmt.Errorf("failed to merge config structs: %w", err)
			}
			log.Debug().
				Msgf("After merge: %+v", conf)
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

// Generic function to marshal and unmarshal configuration
func marshalUnmarshalConfig[T any](configRaw map[string]interface{}, fieldName string, target *T) error {
	fieldData, exists := configRaw[fieldName]
	if !exists {
		return fmt.Errorf("field %s not found in configuration", fieldName)
	}

	fieldYAML, err := yaml.Marshal(fieldData)
	if err != nil {
		return fmt.Errorf("failed to marshal %s config: %w", fieldName, err)
	}

	if err = yaml.Unmarshal(fieldYAML, target); err != nil {
		return fmt.Errorf("failed to unmarshal %s config: %w", fieldName, err)
	}

	return nil
}

func parseRaw(configLocation string) (conf *Config, err error) {
	conf = &Config{}

	configRaw, err := parser.Parse(configLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	err = mapstructure.Decode(configRaw, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	log.Debug().Msgf("conf: %v", conf)

	// convert the resulting bulk map into singular key-value pairs
	entries := lo.Entries(configRaw)
	log.Debug().Msgf("entries: %v", entries)

	// Define a mapping of field names to their corresponding struct pointers
	fieldMappings := map[string]interface{}{
		"database":    &conf.DBConfig, // assuming it's a slice
		"fiber":       &conf.Fiber,
		"signing_key": &conf.SigningKey,
		"secret":      &conf.Secret,
		"librate_env": &conf.LibrateEnv,
	}

	for fieldName, target := range fieldMappings {
		fieldEntries := make(map[string]interface{})

		for i := range entries {
			if strings.HasPrefix(entries[i].Key, fieldName+".") {
				subKey := strings.TrimPrefix(entries[i].Key, fieldName+".")
				fieldEntries[subKey] = entries[i].Value
			}

			if len(fieldEntries) > 0 {
				if err := marshalUnmarshalConfig(fieldEntries, "", &target); err != nil {
					return nil, err
				}
			} else {
				log.Debug().Msgf("no entries found for %s", fieldName)
			}
		}
	}

	log.Info().Msgf("got config: %v", configRaw)

	configStr := createKVPairs(configRaw)
	log.Debug().Msgf("configStr: %s", configStr)
	mapped := config.MapStruct(configStr, conf)
	log.Debug().Msgf("mapped: %v", mapped)
	log.Debug().Msgf("conf: %v", conf)

	return conf, nil
}
