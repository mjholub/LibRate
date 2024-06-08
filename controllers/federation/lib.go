package federation

import (
	"context"

	"github.com/go-ap/activitypub"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/models/member"
)

// FedHandler is the interface for the federation handler
type FedHandler interface {
	Converter
	Follow(c *fiber.Ctx) error
	In(c *fiber.Ctx) error
	Unknown(c *fiber.Ctx) error
}

type Converter interface {
	// FollowToAS converts LibRate FollowBlockRequest to ActivityPub Follow
	FollowToAS(ctx context.Context, req *member.FollowBlockRequest) (*activitypub.Follow, error)
	MemberToActor(c *fiber.Ctx, m *member.Member) ([]byte, error)
}

// FedController holds the dependencies for the federation handler
type FedController struct {
	log     *zerolog.Logger
	storage *pgxpool.Pool
	members member.Storer
	Converter
}

type ConversionHandler struct {
	log *zerolog.Logger
}

// NewFedController returns a new FedController
func NewController(log *zerolog.Logger, storage *pgxpool.Pool, memberStorage member.Storer) *FedController {
	return &FedController{log: log, storage: storage, members: memberStorage}
}
