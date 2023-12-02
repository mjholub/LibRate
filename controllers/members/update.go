package members

import (
	"database/sql"
	"strconv"

	"github.com/gofiber/fiber/v2"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

// UpdateMember handles the updating of user information
func (mc *MemberController) Update(c *fiber.Ctx) (err error) {
	mc.log.Info().Msg("Update called")
	var member *member.Member
	err = c.BodyParser(&member)
	if err != nil {
		mc.log.Error().Err(err).Msgf("Error parsing request body: %v", err)
		return h.Res(c, fiber.StatusBadRequest, "Error parsing request body")
	}
	mc.log.Debug().Msgf("member: %+v", member)

	profilePicID := c.Query("profile_pic_id")
	if profilePicID != "" {
		var profilePicIDInt int64
		profilePicIDInt, err = strconv.ParseInt(profilePicID, 10, 64)
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

	mc.log.Debug().Msgf("member (after update): %+v", member)

	err = mc.storage.Update(c.UserContext(), member)
	if err != nil {
		mc.log.Error().Err(err).Msgf("Error updating member: %v", err)
		return h.Res(c, fiber.StatusInternalServerError, "Internal Server Error")
	}
	return h.Res(c, fiber.StatusOK, "success")
}
