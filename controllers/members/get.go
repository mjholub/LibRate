package members

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/middleware"
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
	memberData := member.Member{}
	err := c.BodyParser(&memberData)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Error parsing request body")
	}
	mc.log.Debug().Msgf("Member: %+v", memberData)

	if memberData.MemberName == "" && memberData.Email == "" {
		return h.Res(c, fiber.StatusBadRequest, "No nickname or email provided")
	}

	exists, err := mc.storage.Check(ctx, memberData.Email, memberData.MemberName)
	if err != nil && err != sql.ErrNoRows {
		mc.log.Error().Msgf("Error checking if member \"%s\" exists: %v", memberData.MemberName, err)
		return h.Res(c, fiber.StatusInternalServerError, "Internal Server Error")
	}
	if exists {
		return h.Res(c, fiber.StatusConflict, "not available")
	}
	return h.Res(c, fiber.StatusOK, "available")
}

// @Summary Get a member (user) by nickname or email
// @Description Retrieve the information the requester is allowed to see about a member
// @Tags accounts,interactions,metadata
// @Param email_or_username path string true "The nickname or email of the member to get"
// @Accept json application/activity+json
// @Success 200 {object} h.ResponseHTTP{data=member.Member}
// @Failure 401 {object} h.ResponseHTTP{} "When certain access prerequisites are not met, e.g. a follower's only-visible metadata is requested"
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/{email_or_username}/info [get]
func (mc *MemberController) GetMemberByNickOrEmail(c *fiber.Ctx) error {
	// TODO:
	// 1. compare the requester's public key with the private key in the database
	// 2. if the keys match, proceed with parsing the requester's identity as valid
	// 3. if the keys don't match, check if the member is public
	// 4. by default, we fall back to noauth

	authorized := c.Request().Header.Peek("Authorization")

	if c.Params("email_or_username") == "" {
		return h.Res(c, fiber.StatusNotFound, "No email or nickname provided")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	memberData, err := mc.storage.Read(ctx, c.Params("email_or_username"), "nick", "email")
	if err != nil {
		mc.log.Error().Msgf("Error getting member \"%s\": %v", c.Params("email_or_username"), err)
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	mc.log.Info().Msgf("Member: %+v", memberData)
	// check for authorization and if the request was made by a non-authorized user and the member.Visibility is not public, return 401
	if len(authorized) == 0 && memberData.Visibility != "public" {
		return h.Res(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	if memberData.Visibility != "public" {
		sessionData, err := mc.sessionStore.Get(c)
		if err != nil {
			return h.Res(c, fiber.StatusInternalServerError, "Error retrieving session")
		}

		token, err := middleware.DecryptJWT(string(authorized), sessionData, mc.conf)
		if err != nil {
			return h.Res(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		viewable, err := mc.canView(c.UserContext(), token, memberData.Webfinger)
		if err != nil {
			return h.Res(c, fiber.StatusInternalServerError, "Error verifying viewability")
		}
		if !viewable {
			return h.Res(c, fiber.StatusUnauthorized, "Unauthorized")
		}
	}

	if memberData.ProfilePicID.Valid {
		memberData.ProfilePicSource, err = mc.images.GetImageSource(c.UserContext(), memberData.ProfilePicID.Int64)
		if err != nil {
			mc.log.Warn().Msgf(
				"Error getting profile picture for member \"%s\" despite valid picture ID: %v", c.Params("email_or_username"), err)
			// send a warning in headers
			c.Set("X-Warning", "Error getting profile picture for member")
			return c.SendStatus(fiber.StatusOK)
		}
	}

	return h.ResData(c, fiber.StatusOK, "success", memberData)
}

// TODO: add webfinger to database
func (mc *MemberController) GetMemberByWebfinger(c *fiber.Ctx) error {
	mc.log.Info().Msg("GetMemberByWebfinger called")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mc.log.Debug().Msgf("Webfinger: %s", c.Params("webfinger"))
	memberData, err := mc.storage.Read(ctx, "webfinger", c.Params("webfinger"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	mc.log.Info().Msgf("Member: %+v", memberData)

	return h.ResData(c, fiber.StatusOK, "success", memberData)
}

func (mc *MemberController) GetID(c *fiber.Ctx) error {
	mc.log.Info().Msg("GetID called")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mc.log.Debug().Msgf("ID: %s", c.Params("id"))
	memberData, err := mc.storage.Read(ctx, "id", c.Params("id"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	mc.log.Info().Msgf("Member: %+v", memberData)

	return h.ResData(c, fiber.StatusOK, "success", memberData)
}

func (mc *MemberController) canView(ctx context.Context, authorization *jwt.Token, viewee string) (bool, error) {
	if viewee == "" {
		return false, fmt.Errorf("No nickname or email provided")
	}

	viewer := authorization.Claims.(jwt.MapClaims)["webfinger"].(string)
	mc.log.Debug().Msgf("Viewer: %s", viewer)

	canView, err := mc.storage.VerifyViewability(ctx, viewer, viewee)
	if err != nil {
		mc.log.Log().Err(err).Msgf("Error verifying viewability of member \"%s\": %v", viewee, err)
		return false, err
	}
	return canView, nil
}
