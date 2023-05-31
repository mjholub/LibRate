package cfg

import (
	"fmt"
	"os"
	"path/filepath"

	"codeberg.org/mjh/LibRate/cfg/parser"
	"codeberg.org/mjh/LibRate/internal/clitools"
	"codeberg.org/mjh/LibRate/internal/logging"
	"gopkg.in/yaml.v3"

	config "github.com/gookit/config/v2"
	"github.com/imdario/mergo"
	"github.com/samber/lo"
)

// nolint:gochecknoglobals
var log = logging.Init()

func LoadConfig() (Config, error) {
	// TODO: parallelize looping over config locations
	// i.e. use some queue to send the config locations to a goroutine pool
	// and return the first config location that is found
	locs := tryLocations()
	loc := lookForExisting(locs)
	if loc == "" {
		userSpecifiedLoc, err := tryGettingConfig(locs)
		if err != nil {
			// FIXME: remove deep exits?
			log.Panic().Err(err).Msgf("error getting config: %s", err.Error())
		}
		if lookForExisting([]string{userSpecifiedLoc}) == "" {
			err := writeConfig(userSpecifiedLoc, Config{})
			if err != nil {
				log.Panic().Err(err).Msgf("error writing config: %s", err.Error())
			}
			log.Info().Msgf("wrote config to %s", userSpecifiedLoc)
		}
		// use the user-specified config location if it exists
		loc = userSpecifiedLoc
	}
	log.Info().Msgf("using config file %s", loc)
	configRaw, err := parser.Parse(loc)
	if err != nil {
		log.Panic().Err(err).Msgf("error parsing config: %s", err.Error())
	}
	log.Info().Msgf("got config: %v", configRaw)

	conf := Config{}
	conf = config.MapStructure(configRaw, conf).(Config)

	if err := mergo.Merge(&conf, readDefaults()); err != nil {
		return conf, err
	}

	return conf, nil
}

func readDefaults() Config {
	return Config{
		DBConfig: DBConfig{
			Engine:   "postgres",
			Host:     "localhost",
			Port:     uint16(5432),
			Database: "librerym",
			User:     "postgres",
			Password: "postgres",
		},
		Fiber: FiberConfig{
			Host: "localhost",
			Port: "3000",
		},
		// Set your default values for SiginingKey and DBPass here.
	}
}

func tryLocations() []string {
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
	configLocations = lo.FlatMap(configLocations, func(s string) []string {
		return lo.Map(configExtensions, func(s2 string) string {
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
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating config dir: %w", err)
	}
	err = os.WriteFile(configPath, yaml, 0640)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
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
