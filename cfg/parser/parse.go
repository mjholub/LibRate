package parser

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"
)

func Parse(filename string) (kv map[string]interface{}, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	err = yaml.Unmarshal(data, &kv)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return kv, nil
}
