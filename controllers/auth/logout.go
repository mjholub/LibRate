package auth

import (
	"github.com/gofiber/fiber/v2"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

func (a *Service) Logout(c *fiber.Ctx) error {
	a.log.Debug().Msg("Logout request")
	session, err := a.sess.Get(c)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, err.Error())
	}
	err = a.sess.Delete(session.ID())
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, err.Error())
	}
	c.ClearCookie("session")
	return c.SendStatus(fiber.StatusOK)
}
