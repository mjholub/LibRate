package cfg

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"codeberg.org/mjh/LibRate/cfg/parser"
	"codeberg.org/mjh/LibRate/internal/clitools"
	"codeberg.org/mjh/LibRate/internal/logging"

	config "github.com/gookit/config/v2"
	"github.com/imdario/mergo"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// nolint:gochecknoglobals
var log = logging.Init()

func LoadConfig() mo.Result[Config] {
	// TODO: parallelize looping over config locations
	// i.e. use some queue to send the config locations to a goroutine pool
	// and return the first config location that is found
	return mo.Try(func() (Config, error) {
		locs := tryLocations()
		loc := lookForExisting(locs)
		if loc == "" {
			userSpecifiedLoc, err := tryGettingConfig(locs)
			if err != nil {
				panic(fmt.Errorf("failed to get config: %w", err))
			}
			if lookForExisting([]string{userSpecifiedLoc}) == "" {
				err := writeConfig(userSpecifiedLoc, Config{})
				if err != nil {
					panic(fmt.Errorf("failed to write config: %w", err))
				}
			}
			loc = userSpecifiedLoc
		}
		configRaw, err := parser.Parse(loc)
		if err != nil {
			panic(fmt.Errorf("failed to parse config: %w", err))
		}
		log.Info().Msgf("got config: %v", configRaw)

		// preallocate default config
		conf := Config{
			DBConfig: DBConfig{
				Engine:   "postgres",
				Host:     "localhost",
				Port:     uint16(5432),
				Database: "librate",
				User:     "postgres",
				Password: "postgres",
				SSL:      "disable",
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
			Engine:   "postgres",
			Host:     "localhost",
			Port:     uint16(5432),
			Database: "librerym",
			User:     "postgres",
			Password: "postgres",
			SSL:      "disable",
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
	if configFileEnv := os.Getenv("CONFIG_FILE"); configFileEnv != "" {
		if _, err := os.Stat(configFileEnv); err == nil {
			return configFileEnv
		}
		return "" // WARN: is this a correct branch?
	}
	for i := range configLocations {
		if _, err := os.Stat(configLocations[i]); err == nil {
			return configLocations[i]
		}
	}
	return ""
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
func writeConfig(configPath string, c Config) error {
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

func LoadDgraph() *DgraphConfig {
	var (
		dghost        string
		dgport        string
		dghttp        string
		dgAlphaBadger string
		dgAlphaBRate  string
		dgAlphaTrace  string
		dgAlphaTLS    string
		dgAlphaSec    string
	)

	envChan := make(chan string, 1)
	defer close(envChan)

	getEnvOrDefault := func(envVar, defaultValue string) string {
		value := os.Getenv(envVar)
		if value == "" {
			os.Setenv(envVar, defaultValue)
			value = defaultValue
		}
		envChan <- value
		return value
	}
	go func() {
		dghost = getEnvOrDefault("DGRAPH_HOST", "0.0.0.0")
		dgport = getEnvOrDefault("DGRAPH_GRPC_PORT", "5080")
		dghttp = getEnvOrDefault("DGRAPH_HTTP_PORT", "6080")
		dgAlphaBadger = getEnvOrDefault("DGRAPH_ALPHA_BADGER", "compression=zstd;cache_size=1G;cache_ttl=1h;max_table_size=1G;level_size=128MB")
		dgAlphaBRate = getEnvOrDefault("DGRAPH_ALPHA_BLOCK_RATE", "20")
		dgAlphaTrace = getEnvOrDefault("DGRAPH_ALPHA_TRACE", "prometheus=localhost:9090")
		dgAlphaTLS = getEnvOrDefault("DGRAPH_ALPHA_TLS", "false")
		dgAlphaSec = getEnvOrDefault("DGRAPH_ALPHA_SECURITY", `whitelist=
		10.0.0.0/8,
		172.0.0.0/8,
		192.168.0.0/16,
		`+dghost+`
		`)
	}()

	// Retrieve the values from the channel
	dghost = <-envChan
	dgport = <-envChan
	dghttp = <-envChan
	dgAlphaBadger = <-envChan
	dgAlphaBRate = <-envChan
	dgAlphaTrace = <-envChan
	dgAlphaTLS = <-envChan
	dgAlphaSec = <-envChan

	return &DgraphConfig{
		Host:           dghost,
		GRPCPort:       dgport,
		HTTPPort:       dghttp,
		AlphaBadger:    dgAlphaBadger,
		AlphaBlockRate: dgAlphaBRate,
		AlphaTrace:     dgAlphaTrace,
		AlphaTLS:       dgAlphaTLS,
		AlphaSecurity:  dgAlphaSec,
	}
}

func createKVPairs(m map[string]interface{}) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}
