package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"

	"codeberg.org/mjh/LibRate/cfg"
)

// Protected protect routes
func Protected() fiber.Handler {
	conf, err := cfg.LoadConfig().Get()
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
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT", "data": nil})
	}
	return c.Status(fiber.StatusUnauthorized).
		JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT", "data": nil})
}
