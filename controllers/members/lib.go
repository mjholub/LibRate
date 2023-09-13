package members

import (
	"codeberg.org/mjh/LibRate/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// MemberController allows for the retrieval of user information
type (
	// IMemberController is the interface for the member controller
	// It defines the methods that the member controller must implement
	// This is useful for mocking the member controller in unit tests
	IMemberController interface {
		GetMember(c *fiber.Ctx) error
		UpdateMember(c *fiber.Ctx) error
		DeleteMember(c *fiber.Ctx) error
	}

	// MemberController is the controller for member endpoints
	MemberController struct {
		storage *models.MemberStorage
		log     *zerolog.Logger
	}
)

func NewController(storage *models.MemberStorage, logger *zerolog.Logger) *MemberController {
	return &MemberController{storage: storage, log: logger}
}
