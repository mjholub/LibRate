package render

import (
	"testing"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/internal/logging"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	conf := &cfg.Config{
		LibrateEnv: "development",
		Logging: logging.Config{
			Level: "trace",
		},
	}

	eng := Setup(conf)
	assert.NotNilf(t, eng, "expected engine to be not nil")
	assert.NotNilf(t, eng.Templates, "expected engine templates to be not nil")
}
