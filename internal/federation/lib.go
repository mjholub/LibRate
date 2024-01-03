package federation

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/models/member"
)

// FedHandler is the interface for the federation handler
type FedHandler interface {
	Follow(c *fiber.Ctx) error
	AcceptFollow(c *fiber.Ctx) error
	RejectFollow(c *fiber.Ctx) error
	Unfollow(c *fiber.Ctx) error
}

// FedController holds the dependencies for the federation handler
type FedController struct {
	log     zerolog.Logger
	storage *sqlx.DB
	members member.MemberStorer
}

// NewFedController returns a new FedController
func NewFedController(log zerolog.Logger, storage *sqlx.DB) *FedController {
	return &FedController{log: log, storage: storage}
}
