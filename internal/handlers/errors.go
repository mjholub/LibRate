package handlers

import "github.com/gofiber/fiber/v2"

func Res(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"message": message,
	})
}
