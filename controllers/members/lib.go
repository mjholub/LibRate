package members

import (
	"github.com/go-ap/activitypub"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
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
		storage member.Storer
		log     *zerolog.Logger
		conf    *cfg.Config
		images  *static.Storage
	}
)

func NewController(
	storage member.Storer,
	db *sqlx.DB,
	logger *zerolog.Logger,
	conf *cfg.Config,
) *MemberController {
	imagesStorage := static.NewStorage(db, logger)
	return &MemberController{storage: storage, log: logger, conf: conf, images: imagesStorage}
}

// MemberToActor converts a member to an ActivityPub actor
func MemberToActor(c *fiber.Ctx, member *member.Member) ([]byte, error) {
	base := c.BaseURL() + "/api/members/" + member.MemberName
	return activitypub.Actor{
		ID:        activitypub.IRI(c.BaseURL() + "/api/members/" + member.MemberName),
		Type:      activitypub.PersonType,
		Inbox:     activitypub.IRI(base + "/inbox"),
		Outbox:    activitypub.IRI(base + "/outbox"),
		Following: activitypub.IRI(base + "/following"),
		Followers: activitypub.IRI(base + "/followers"),
		Liked:     activitypub.IRI(base + "/liked"),
		PreferredUsername: activitypub.NaturalLanguageValues{
			activitypub.DefaultLangRef(member.DisplayName.String),
		},
		Endpoints: &activitypub.Endpoints{
			SharedInbox: activitypub.IRI(c.BaseURL() + "/api/inbox"),
		},
		PublicKey: activitypub.PublicKey{
			ID:           activitypub.IRI(base + "#main-key"),
			Owner:        activitypub.IRI(base),
			PublicKeyPem: member.PublicKeyPem,
		},
	}.GobEncode()
}
