package main

import (
	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()
	app.Use(logger.New())
	cfg.LoadConfig()

	routes.Setup(app)

	err := app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
