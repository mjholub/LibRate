package middleware

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// Protected protect routes
func Protected(log *zerolog.Logger, conf *cfg.Config) fiber.Handler {
	if log != nil {
		log.Debug().Msgf("Protected middleware: Signing key: %v", conf.Secret)
	}
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(conf.Secret)},
		ErrorHandler: jwtError,
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	if err.Error() == "Missing or malformed JWT" {
		return h.ResData(c, fiber.StatusBadRequest, "Missing or malformed JWT", nil)
	}
	return h.ResData(c, fiber.StatusUnauthorized, "Invalid or expired JWT: "+err.Error(), nil)
}
