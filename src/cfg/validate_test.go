package cfg

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	conf := TestConfig
	validate := validator.New()
	errs := Validate(&conf, validate)
	assert.Empty(t, errs)
}
