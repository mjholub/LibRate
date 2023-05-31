package parser

import (
	config "github.com/gookit/config/v2"
	yaml "github.com/gookit/config/v2/yaml"
)

func Parse(filename string) (kv map[string]interface{}, err error) {
	config.WithOptions(config.ParseEnv) // optional: enable environment variable parsing
	config.AddDriver(yaml.Driver)       // add YAML driver

	// load from file
	if err := config.LoadFiles(filename); err != nil {
		return kv, err
	}

	if err := config.Scan(&kv); err != nil {
		return kv, err
	}

	return kv, nil
}
