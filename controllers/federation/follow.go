package federation

import (
	"fmt"

	"github.com/go-ap/activitypub"
	"github.com/gofiber/fiber/v2"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

// In handles incoming messages, then routes them to the appropriate handler
func (fc *FedController) In(c *fiber.Ctx) error {
	var activity activitypub.Activity

	if err := c.BodyParser(&activity); err != nil {
		return err
	}

	switch activity.Type {
	case activitypub.FollowType:
		return fc.Follow(c)
	default:
		return fc.Unknown(c)
	}
}

// Unknown handles unknown activity types
func (fc *FedController) Unknown(c *fiber.Ctx) error {
	return h.Res(c, fiber.StatusNotImplemented, "Unknown activity type")
}

// Follow handles incoming follow requests
func (fc *FedController) Follow(c *fiber.Ctx) error {
	var activity activitypub.Follow
	if err := c.BodyParser(&activity); err != nil {
		return err
	}
	actor := activity.Actor.GetLink()
	follows := activity.Object.GetLink()
	fc.log.Info().Msgf("Follow request from %s to %s", actor, follows)

	err := fc.members.RequestFollow(c.Context(), &member.FollowBlockRequest{
		Requester: actor.String(),
		Target:    follows.String(),
	})
	if err != nil {
		return fmt.Errorf("error requesting follow: %w", err)
	}

	return h.Res(c, fiber.StatusOK, "Follow request sent")
}
