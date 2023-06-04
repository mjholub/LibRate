package main

import (
	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"codeberg.org/mjh/LibRate/internal/logging"
)

func main() {
	app := fiber.New()
	app.Use(logger.New())
	conf, err := cfg.LoadConfig()
	log := logging.Init()
	if err != nil {
		log.Panic().Err(err).Msg("Failed to load config")
	}
	// CORS
	app.Use("/api", func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", conf.Host)
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		return c.Next()
	})

	routes.Setup(app)

	err = app.Listen(":3000")
	if err != nil {
		log.Panic().Err(err).Msg("Failed to listen on port 3000")
	}
}
