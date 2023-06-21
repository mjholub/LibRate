package cfg

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"codeberg.org/mjh/LibRate/cfg/parser"
	"codeberg.org/mjh/LibRate/internal/clitools"
	"codeberg.org/mjh/LibRate/internal/logging"

	config "github.com/gookit/config/v2"
	"github.com/imdario/mergo"
	"github.com/reactivex/rxgo/v2"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// nolint:gochecknoglobals
var log = logging.Init()

func LoadConfig() mo.Result[Config] {
	return mo.Try(func() (Config, error) {
		pg_config, err := exec.LookPath("pg_config")
		if err != nil {
			return Config{}, fmt.Errorf("failed to find pg_config: %w. Is postgres installed?", err)
		}
		if os.Getenv("LIBRATE_ENV") == "test" {
			return Config{
				DBConfig: DBConfig{
					Engine:    "postgres",
					Host:      "localhost",
					Port:      uint16(5432),
					Database:  "librate_test",
					User:      "postgres",
					Password:  "postgres",
					SSL:       "disable",
					PG_Config: pg_config,
				},
				Fiber: FiberConfig{
					Host: "localhost",
					Port: "3000",
				},
				SiginingKey: "",
				DBPass:      "postgres",
			}, nil
		}
		locs := tryLocations()
		loc := lookForExisting(locs)
		if loc == "" {
			userSpecifiedLoc, err := tryGettingConfig(locs)
			if err != nil {
				return Config{}, fmt.Errorf("failed to get config: %w", err)
			}
			if lookForExisting([]string{userSpecifiedLoc}) == "" {
				err := writeConfig(userSpecifiedLoc, &Config{})
				if err != nil {
					return Config{}, fmt.Errorf("failed to write config: %w", err)
				}
			}
			loc = userSpecifiedLoc
		}
		configRaw, err := parser.Parse(loc)
		if err != nil {
			return Config{}, fmt.Errorf("failed to parse config: %w", err)
		}
		log.Info().Msgf("got config: %v", configRaw)

		// preallocate default config
		conf := Config{
			DBConfig: DBConfig{
				Engine:    "postgres",
				Host:      "localhost",
				Port:      uint16(5432),
				Database:  "librate",
				User:      "postgres",
				Password:  "postgres",
				SSL:       "disable",
				PG_Config: pg_config,
			},
			Fiber: FiberConfig{
				Host: "localhost",
				Port: "3000",
			},
			SiginingKey: "",
			DBPass:      "postgres",
		}

		configStr := createKVPairs(configRaw)
		_ = config.MapStruct(configStr, conf) // WARN: unsure if this is correct

		if err := mergo.Merge(&conf, ReadDefaults()); err != nil {
			return conf, fmt.Errorf("failed to merge config structs: %w", err)
		}

		return conf, nil
	})
}

func ReadDefaults() Config {
	return Config{
		DBConfig: DBConfig{
			Engine:    "postgres",
			Host:      "localhost",
			Port:      uint16(5432),
			Database:  "librerym",
			User:      "postgres",
			Password:  "postgres",
			SSL:       "disable",
			PG_Config: "/usr/bin/pg_config",
		},
		Fiber: FiberConfig{
			Host: "localhost",
			Port: "3000",
		},
		// Set your default values for SiginingKey and DBPass here.
	}
}

func tryLocations() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Panic().Err(err).Msgf("error getting user home dir: %s", err.Error())
	}
	cfhome, err := os.UserConfigDir()
	if err != nil || cfhome == "" {
		if _, err := os.Stat(filepath.Join(home, ".config")); err == nil {
			cfhome = filepath.Join(home, ".config")
		} else {
			log.Panic().Err(err).Msgf("error getting user config dir: %s", err.Error())
		}
	}

	configLocations := []string{
		"config",
		"config/config",
		"/etc/librate/config",
		"/var/lib/librate/config",
		"/opt/librate/config",
		"/usr/local/librate/config",
		home + "/.config/librate/config",
		cfhome + "/.local/share/librate/config",
	}
	configExtensions := []string{
		".yml",
		".yaml",
		"",
	}
	// use FlatMap and Map to create a list of all possible config file locations
	configLocations = lo.FlatMap(configLocations, func(s string, _ int) []string {
		return lo.Map(configExtensions, func(s2 string, _ int) string {
			return s + s2
		})
	})
	return configLocations
}

func lookForExisting(configLocations []string) string {
	// create a channel to emit the file pahts
	fpChan := make(chan rxgo.Item)

	// WaitGroup to ensure all checks are done before closing fpChan
	var wg sync.WaitGroup

	if configFileEnv := os.Getenv("CONFIG_FILE"); configFileEnv != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := os.Stat(configFileEnv); err == nil {
				fpChan <- rxgo.Of(configFileEnv)
			}
		}()
	}

	for i := range configLocations {
		wg.Add(1)
		path := configLocations[i]
		go func(i int) {
			defer wg.Done()
			if _, err := os.Stat(configLocations[i]); err == nil {
				fpChan <- rxgo.Of(path)
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(fpChan)
	}()

	// create an observable from the channel
	fpObs := rxgo.FromChannel(fpChan)

	item, err := fpObs.First().Get()
	if err != nil {
		return "" // no config file found
	}
	// return the first config file found
	return item.V.(string)
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

func getDefaultConfigPath() (string, error) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("error getting user config dir: %w", err)
	}
	defaultConfigPath := filepath.Join(confDir, "librate", "config.yaml")
	return defaultConfigPath, nil
}

// TODO: also create .env file in the same dir with CONFIG_FILE set to the path to make looking
// up the config file faster in the future
func writeConfig(configPath string, c *Config) error {
	if configPath == "" {
		return fmt.Errorf("no config file specified")
	}
	yaml, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}
	configDir := filepath.Dir(configPath)
	err = os.MkdirAll(configDir, 0o755)
	if err != nil {
		return fmt.Errorf("error creating config dir: %w", err)
	}
	err = os.WriteFile(configPath, yaml, 0o640)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	return nil
}

func createKVPairs(m map[string]interface{}) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}
