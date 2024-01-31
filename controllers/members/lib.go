package members

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers/federation"
	"codeberg.org/mjh/LibRate/models/member"
	"codeberg.org/mjh/LibRate/models/static"
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
		fedCon       federation.FedHandler
		storage      member.Storer
		sessionStore *session.Store
		log          *zerolog.Logger
		conf         *cfg.Config
		images       *static.Storage
	}
)

func NewController(
	storage member.Storer,
	db *sqlx.DB,
	sess *session.Store,
	logger *zerolog.Logger,
	conf *cfg.Config,
) *MemberController {
	imagesStorage := static.NewStorage(db, logger)
	return &MemberController{
		storage:      storage,
		fedCon:       federation.NewController(logger, db, storage),
		sessionStore: sess,
		log:          logger,
		conf:         conf,
		images:       imagesStorage,
	}
}
