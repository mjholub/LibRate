package federation

import (
	"context"
	"fmt"

	"github.com/go-ap/activitypub"
	"github.com/gofiber/fiber/v2"

	"codeberg.org/mjh/LibRate/models/member"
)

// MemberToActor converts a member to an ActivityPub actor
func (ch *ConversionHandler) MemberToActor(c *fiber.Ctx, memberData *member.Member) ([]byte, error) {
	// TODO: add multilingual support
	base := c.BaseURL() + "/api/members/" + memberData.MemberName
	actor, err := activitypub.Actor{
		ID:        activitypub.IRI(c.BaseURL() + "/api/members/" + memberData.MemberName),
		Type:      activitypub.PersonType,
		Inbox:     activitypub.IRI(base + "/inbox"),
		Outbox:    activitypub.IRI(base + "/outbox"),
		Following: activitypub.IRI(base + "/following"),
		Followers: activitypub.IRI(base + "/followers"),
		Liked:     activitypub.IRI(base + "/liked"),
		PreferredUsername: []activitypub.LangRefValue{
			{
				Ref:   "eng",
				Value: activitypub.Content(memberData.DisplayName.String),
			},
		},
		Endpoints: &activitypub.Endpoints{
			SharedInbox: activitypub.IRI(c.BaseURL() + "/api/inbox"),
		},
		PublicKey: activitypub.PublicKey{
			ID:           activitypub.IRI(base + "#main-key"),
			Owner:        activitypub.IRI(base),
			PublicKeyPem: memberData.PublicKeyPem,
		},
	}.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error encoding actor: %v", err)
	}
	return actor, nil
}

// TODO: create an actual implementation
func (ch *ConversionHandler) FollowToAS(ctx context.Context, req *member.FollowBlockRequest) (*activitypub.Follow, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &activitypub.Follow{
			Actor:  activitypub.IRI(req.Requester),
			Object: activitypub.IRI(req.Target),
		}, nil
	}
}
