package cfg

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/samber/lo"
)

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
		home + "/.local/share/librate/config",
		cfhome + "librate/config",
	}
	configExtensions := []string{
		".yml",
		".yaml",
		"_enc.yml",
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
func lookForExisting(configLocations []string) (string, error) {
	// Create a channel for emitting file paths
	resultChan := make(chan struct {
		Path string
		Err  error
	})

	go func() {
		// Check the CONFIG_FILE environment variable
		if configFileEnv := os.Getenv("CONFIG_FILE"); configFileEnv != "" {
			if _, err := os.Stat(configFileEnv); err == nil {
				resultChan <- struct {
					Path string
					Err  error
				}{configFileEnv, nil}
				return
			}
		}

		// Check the other config locations
		for _, path := range configLocations {
			if _, err := os.Stat(path); err == nil {
				resultChan <- struct {
					Path string
					Err  error
				}{path, nil}
				return
			}
		}

		// If no file is found, emit an error
		resultChan <- struct {
			Path string
			Err  error
		}{"", errors.New("no config file found")}
	}()

	// Extract the result from the channel
	result := <-resultChan

	return result.Path, result.Err
}

// FileExists checks whether the config file exists. It is useful for the fallback mechanism of using default config
func FileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
