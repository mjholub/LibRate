package members

import (
	"database/sql"
	"strconv"

	"github.com/gofiber/fiber/v2"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

// UpdateMember handles the updating of user information
func (mc *MemberController) Update(c *fiber.Ctx) error {
	mc.log.Info().Msg("Update called")
	var member *member.Member
	err := c.BodyParser(&member)
	if err != nil {
		mc.log.Error().Err(err).Msgf("Error parsing request body: %v", err)
		return h.Res(c, fiber.StatusBadRequest, "Error parsing request body")
	}
	// NOTE: this is a hack. We're passing the profile pic ID in the
	// request body, but c.BodyParser() doesn't include it, because we don't
	// want the ID in the JSON response. So we're setting it here.
	profilePicID := c.Params("profilePic")
	if profilePicID != "" {
		profilePicIDInt, err := strconv.ParseInt(profilePicID, 10, 64)
		if err != nil {
			mc.log.Error().Err(err).Msgf("Error parsing profile pic ID: %v", err)
			return h.Res(c, fiber.StatusBadRequest, "Error parsing profile pic ID")
		}
		member.ProfilePicID = sql.NullInt64{
			Int64: profilePicIDInt,
			Valid: true,
		}
		mc.log.Debug().Msgf("Profile pic ID: %v", member.ProfilePicID)
	}

	err = mc.storage.Update(c.UserContext(), member)
	if err != nil {
		mc.log.Error().Err(err).Msgf("Error updating member: %v", err)
		return h.Res(c, fiber.StatusInternalServerError, "Internal Server Error")
	}
	return h.Res(c, fiber.StatusOK, "success")
}
