package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type ResponseHTTP struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Res(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"message": message,
	})
}

func ResData(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"message": message,
		"data":    data,
	})
}

func InternalError(log *zerolog.Logger, c *fiber.Ctx, message string, err error) error {
	log.Error().Err(err).Msg(message)
	return Res(c, fiber.StatusInternalServerError, message)
}

func BadRequest(
	log *zerolog.Logger,
	c *fiber.Ctx,
	clientMessage, serverMessage string,
	err error,
) error {
	log.Error().Msgf("Failed to %s: %v", serverMessage, err)
	return Res(c, fiber.StatusBadRequest, clientMessage)
}
