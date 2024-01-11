package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"codeberg.org/mjh/LibRate/cfg"
)

func TestDBRunning(t *testing.T) {
	assert := assert.New(t)
	conf := &cfg.TestConfig.DBConfig
	assert.True(DBRunning(conf.Port))
	assert.False(DBRunning(conf.Port + 100))
}

// test if flags are properly parsed and assigned their default values

func TestParseFlags(t *testing.T) {
	assert := assert.New(t)
	flags := parseFlags()
	assert.False(flags.Exit)
}
