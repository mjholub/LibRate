package cfg_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPGConfigPath(t *testing.T) {
	pgConfig, err := exec.LookPath("pg_config")
	assert.NoError(t, err)
	assert.NotEmpty(t, pgConfig)
}
