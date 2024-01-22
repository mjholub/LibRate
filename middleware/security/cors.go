package security

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"codeberg.org/mjh/LibRate/cfg"
)

func SetupCORS(conf *cfg.Config) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, X-Requested-With, X-Csrf-Token",
		AllowMethods: "GET, POST, PUT, DELETE, HEAD, PATCH, OPTIONS",
	})
}
