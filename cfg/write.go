package cfg

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func createKVPairs(m map[string]interface{}) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}

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

	// write the .env file
	envPath := filepath.Join(configDir, ".env")
	env := fmt.Sprintf("CONFIG_FILE=%s", configPath)
	err = os.WriteFile(envPath, []byte(env), 0o640)

	return nil
}
