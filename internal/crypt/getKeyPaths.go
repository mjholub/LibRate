package crypt

import (
	"fmt"
	"os"
	"path/filepath"

	"codeberg.org/mjh/LibRate/cfg"
)

// GetPublicKeyPath returns the absolute path to the public key
func GetPublicKeyPath(conf *cfg.Config) (string, error) {
	return getAbsPathCheckExistence(conf.Keys.Public)
}

func GetPrivateKeyPath(conf *cfg.Config) (string, error) {
	return getAbsPathCheckExistence(conf.Keys.Private)
}

func getAbsPathCheckExistence(path string) (string, error) {
	if string(path[0]) == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get home directory: %w", err)
		}
		path = homeDir + path[1:]
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path to public key: %w", err)
	}
	if _, err := os.Stat(absolutePath); os.IsNotExist(err) {
		return "", fmt.Errorf("public key does not exist or invalid path \"%s\": %w", absolutePath, err)
	}
	return absolutePath, nil
}
