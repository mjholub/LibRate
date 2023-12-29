package middleware

import (
	"os"
	"testing"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/middleware/session"
)

// MockConfig is a mock configuration for testing purposes
type MockConfig struct{}

func (m *MockConfig) Get() mo.Result[*cfg.Config] {
	return mo.Ok(&cfg.Config{})
}

func (m *MockConfig) GetError() mo.Result[cfg.Config] {
	return mo.Err[cfg.Config](nil) // Simulate an error
}

func TestProtected(t *testing.T) {
	store, err := session.Setup(&cfg.TestConfig)
	require.NoError(t, err)
	t.Run("When config is loaded successfully", func(t *testing.T) {
		mockLogger := zerolog.New(os.Stdout)
		handler := Protected(store, &mockLogger, nil)

		// Create a mock Fiber context for testing
		ctx := &fiber.Ctx{}
		err = handler(ctx)

		// Assert that the handler returned no error
		assert.NoError(t, err)
	})

	t.Run("When config fails to load", func(t *testing.T) {
		mockLogger := zerolog.New(os.Stdout)
		handler := Protected(store, &mockLogger, nil)

		// Create a mock Fiber context for testing
		ctx := &fiber.Ctx{}
		err := handler(ctx)

		// Assert that the handler returned an error
		assert.Error(t, err)
	})
}

func TestJwtError(t *testing.T) {
	t.Run("When FIBER_ENV is 'dev' and error is 'Missing or malformed JWT'", func(t *testing.T) {
		os.Setenv("FIBER_ENV", "dev")
		defer os.Unsetenv("FIBER_ENV")

		// Create a mock Fiber context for testing
		ctx := &fiber.Ctx{}
		err := jwtError(ctx, jwtware.ErrJWTMissingOrMalformed)

		// Assert that the handler returned no error
		assert.NoError(t, err)
	})

	t.Run("When FIBER_ENV is not 'dev' and error is 'Missing or malformed JWT'", func(t *testing.T) {
		os.Setenv("FIBER_ENV", "prod")
		defer os.Unsetenv("FIBER_ENV")

		// Create a mock Fiber context for testing
		ctx := &fiber.Ctx{}
		err := jwtError(ctx, jwtware.ErrJWTMissingOrMalformed)

		// Assert that the handler returned an error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Missing or malformed JWT")
	})

	t.Run("When FIBER_ENV is not 'dev' and error is not 'Missing or malformed JWT'", func(t *testing.T) {
		os.Setenv("FIBER_ENV", "prod")
		defer os.Unsetenv("FIBER_ENV")

		// Create a mock Fiber context for testing
		ctx := &fiber.Ctx{}
		err := jwtError(ctx, jwtware.ErrJWTAlg)

		// Assert that the handler returned an error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid or expired JWT")
	})
}
