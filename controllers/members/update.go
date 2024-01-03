package members

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

// UpdateMember handles the updating of user information
func (mc *MemberController) Update(c *fiber.Ctx) (err error) {
	mc.log.Info().Msg("Update called")
	var member *member.Member
	ct := c.Request().Header.Peek("Content-Type")
	if strings.Contains(string(ct), "multipart/form-data") {
		member = parseFormValues(c)
	} else {
		err = c.BodyParser(&member)
		if err != nil {
			mc.log.Error().Err(err).Msgf("Error parsing request body: %v", err)
			return h.Res(c, fiber.StatusBadRequest, "Error parsing request body")
		}
	}
	member.MemberName = c.Params("member_name")
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

// parseFormValues parses the form values from the request body for which member struct fields are present
func parseFormValues(c *fiber.Ctx) (m *member.Member) {
	displayName := c.FormValue("display_name")
	var dn, bioValue, hp sql.NullString
	dn = lo.Ternary(displayName != "", sql.NullString{String: displayName, Valid: true}, sql.NullString{Valid: false})
	email := c.FormValue("email")
	bio := c.FormValue("bio")
	bioValue = lo.Ternary(bio != "", sql.NullString{String: bio, Valid: true}, sql.NullString{Valid: false})
	homepage := c.FormValue("homepage")
	hp = lo.Ternary(homepage != "", sql.NullString{String: homepage, Valid: true}, sql.NullString{Valid: false})
	visibility := c.FormValue("visibility")

	return &member.Member{
		DisplayName: dn,
		Email:       email,
		Bio:         bioValue,
		Homepage:    hp,
		Visibility:  visibility,
		Active:      true,
	}
}
