package cfg

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/reactivex/rxgo/v2"
	"github.com/samber/lo"
)

func getDefaultConfigPath() (string, error) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("error getting user config dir: %w", err)
	}
	defaultConfigPath := filepath.Join(confDir, "librate", "config.yaml")
	return defaultConfigPath, nil
}

// tryLocations returns a list of possible config file locations
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
			log.Fatal().Err(err).Msgf("error getting user config dir: %s", err.Error())
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

// lookForExisting looks for an existing config file in the given locations and returns the first
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
