package members

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

func (mc *MemberController) GetFollowers(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

// check checks for the existence of a member
// it requires both nickname and email to be provided
func (mc *MemberController) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mc.log.Trace().Msgf("Check called with payload: %s", string(c.Request().Body()))
	member := member.Member{}
	err := c.BodyParser(&member)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Error parsing request body")
	}
	mc.log.Debug().Msgf("Member: %+v", member)

	if member.MemberName == "" && member.Email == "" {
		return h.Res(c, fiber.StatusBadRequest, "No nickname or email provided")
	}

	exists, err := mc.storage.Check(ctx, member.Email, member.MemberName)
	if err != nil && err != sql.ErrNoRows {
		mc.log.Error().Msgf("Error checking if member \"%s\" exists: %v", member.MemberName, err)
		return h.Res(c, fiber.StatusInternalServerError, "Internal Server Error")
	}
	if exists {
		return h.Res(c, fiber.StatusConflict, "not available")
	}
	return h.Res(c, fiber.StatusOK, "available")
}

func (mc *MemberController) GetMemberByNickOrEmail(c *fiber.Ctx) error {
	// TODO:
	// 1. compare the requester's public key with the private key in the database
	// 2. if the keys match, proceed with parsing the requester's identity as valid
	// 3. if the keys don't match, check if the member is public
	// 4. by default, we fall back to noauth
	requester := member.Member{} // works like "noauth" in gotosocial

	authorized := c.Request().Header.Peek("Authorization")

	accept := string(c.Request().Header.Peek("Accept"))

	if c.Params("email_or_username") == "" {
		return h.Res(c, fiber.StatusNotFound, "No email or nickname provided")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	member, err := mc.storage.Read(ctx, c.Params("email_or_username"), "nick", "email")
	if err != nil {
		mc.log.Error().Msgf("Error getting member \"%s\": %v", c.Params("email_or_username"), err)
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	// check for authorization and if the request was made by a non-authorized user and the member.Visibility is not public, return 401
	if len(authorized) == 0 && member.Visibility != "public" {
		return h.Res(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	const activityStreams = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""

	var actor []byte

	if accept == activityStreams || strings.HasPrefix(accept, "application/activity+json") {
		actor, err = MemberToActor(c, member)
		if err != nil {
			mc.log.Error().Msgf("Error converting member to actor: %v", err)
			return c.SendStatus(fiber.StatusBadRequest)
		}
		c.Set("Content-Type", "application/activity+json")
		return h.ResData(c, fiber.StatusOK, "success", actor)
	}
	// TODO: check if the requester is a follower when
	// member.Visibility == "followers_only"
	var followStatus bool
	if member.Visibility == "followers_only" {
		followStatus, err = requester.IsFollowing(ctx, member.ID)
		if err != nil {
			// TODO: use webfingers, since MemberName (nick) is bound to current instance
			mc.log.Error().Msgf("Error checking if %s is following %s: %v", requester.MemberName, member.MemberName, err)
			return h.Res(c, fiber.StatusInternalServerError, "Internal Server Error")
		}
		if !followStatus {
			return h.Res(c, fiber.StatusUnauthorized, "Unauthorized")
		}
	}

	if member.ProfilePicID.Valid {
		member.ProfilePicSource, err = mc.images.GetImageSource(c.UserContext(), member.ProfilePicID.Int64)
		if err != nil {
			mc.log.Warn().Msgf(
				"Error getting profile picture for member \"%s\" despite valid picture ID: %v", c.Params("email_or_username"), err)
			// send a warning in headers
			c.Set("X-Warning", "Error getting profile picture for member")
			return c.SendStatus(fiber.StatusOK)
		}
	}
	mc.log.Info().Msgf("Member: %+v", member)

	return h.ResData(c, fiber.StatusOK, "success", member)
}

// TODO: add webfinger to database
func (mc *MemberController) GetMemberByWebfinger(c *fiber.Ctx) error {
	mc.log.Info().Msg("GetMemberByWebfinger called")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mc.log.Debug().Msgf("Webfinger: %s", c.Params("webfinger"))
	member, err := mc.storage.Read(ctx, "webfinger", c.Params("webfinger"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	mc.log.Info().Msgf("Member: %+v", member)

	return h.ResData(c, fiber.StatusOK, "success", member)
}

func (mc *MemberController) GetID(c *fiber.Ctx) error {
	mc.log.Info().Msg("GetID called")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mc.log.Debug().Msgf("ID: %s", c.Params("id"))
	member, err := mc.storage.Read(ctx, "id", c.Params("id"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	mc.log.Info().Msgf("Member: %+v", member)

	return h.ResData(c, fiber.StatusOK, "success", member)
}
