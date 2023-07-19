package middleware

import (
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"

	"codeberg.org/mjh/LibRate/cfg"
	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// Protected protect routes
func Protected() fiber.Handler {
	conf, err := cfg.LoadConfig().Get()
	// this looks ugly, but importing the compact error handler would require changing
	// this function's signature in a way in which it'd accept *fiber.Ctx as a parameter
	if err != nil {
		return func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"status": "error", "message": "Internal Server Error", "data": nil})
		}
	}
	if os.Getenv("FIBER_ENV") == "dev" {
		conf.SigningKey = "dev"
	}
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(conf.SigningKey)},
		ErrorHandler: jwtError,
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return h.ResData(c, fiber.StatusBadRequest, "Missing or malformed JWT", nil)
	}
	return h.ResData(c, fiber.StatusUnauthorized, "Invalid or expired JWT", nil)
}
