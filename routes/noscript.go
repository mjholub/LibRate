package routes

import "github.com/gofiber/fiber/v2"

func SetupNoScript(app *fiber.App) error {
	app.Static("/", "./views")
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./views/index.html")
	})
	app.Get("/profiles/:nick", func(c *fiber.Ctx) error {
		return c.SendFile("./views/profile.html")
	})
	return nil
}
