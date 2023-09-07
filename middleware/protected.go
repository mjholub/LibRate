package middleware

import (
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// Protected protect routes
func Protected(log *zerolog.Logger, conf *cfg.Config) fiber.Handler {
	// this looks ugly, but importing the compact error handler would require changing
	// this function's signature in a way in which it'd accept *fiber.Ctx as a parameter
	if os.Getenv("FIBER_ENV") == "dev" {
		conf.SigningKey = "dev"
		// in most calls we pass nil to avoid spamming the logs with this warning
		if log != nil {
			log.Warn().Msg("JWT signing key is set to 'dev' for development purposes")
		}
	}
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(conf.SigningKey)},
		ErrorHandler: jwtError,
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	if os.Getenv("FIBER_ENV") == "dev" {
		return nil
	}

	if err.Error() == "Missing or malformed JWT" {
		return h.ResData(c, fiber.StatusBadRequest, "Missing or malformed JWT", nil)
	}
	return h.ResData(c, fiber.StatusUnauthorized, "Invalid or expired JWT", nil)
}
