package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBRunning(t *testing.T) {
	assert := assert.New(t)
	assert.True(DBRunning(false, 5432))
	assert.False(DBRunning(true, 5433))
	assert.True(DBRunning(true, 5433))
}

// test if flags are properly parsed and assigned their default values

func TestParseFlags(t *testing.T) {
	assert := assert.New(t)
	flags := parseFlags()
	assert.False(flags.Exit)
}
