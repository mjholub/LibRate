package parser

import (
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

func Parse(filename string) (kv map[string]interface{}, err error) {
	_ = config.NewWithOptions("conf", config.ParseEnv)
	config.AddDriver(yaml.Driver) // add YAML driver

	// load from file
	if err := config.LoadFiles(filename); err != nil {
		return kv, err
	}

	kv = config.Data()

	return kv, nil
}
