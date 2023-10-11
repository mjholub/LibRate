package members

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models/member"
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
		storage member.MemberStorer
		log     *zerolog.Logger
		conf    *cfg.Config
	}
)

func NewController(
	storage member.MemberStorer,
	logger *zerolog.Logger,
	conf *cfg.Config,
) *MemberController {
	return &MemberController{storage: storage, log: logger, conf: conf}
}
