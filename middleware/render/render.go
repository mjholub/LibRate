package render

import (
	"github.com/gofiber/fiber/v2"
)

// Render renders the template that corresponds to request path
func Render(c *fiber.Ctx) error {
	// get the path from the request
	path := c.Path()
	// render the template corresponding to the path
	// TODO: use the actual template
	return c.Render(path, fiber.Map{})
}
