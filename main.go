package main

import (
	"librerym/cfg"
	"librerym/utils"

	"github.com/gofiber/fiber/v2"
)

func main() {
	log := utils.NewLogger()
	log.Info("Starting server...")
	app := fiber.New()
	cfg.LoadConfig()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./fe/public/index.html")
	})
	app.Listen(":3000")
}
