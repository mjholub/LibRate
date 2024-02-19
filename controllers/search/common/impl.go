package common

import "github.com/gofiber/fiber/v2"

type Searcher interface {
	HandleSearch(c *fiber.Ctx) error
}
