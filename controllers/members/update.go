package members

import (
	"database/sql"
	"fmt"
	"net/mail"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgtype"
	"github.com/lib/pq"
	"github.com/samber/lo"
	"golang.org/x/text/language"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

// UpdateMember handles the updating of user information
// @Summary Update member information
// @Tags accounts,metadata,updating
// @Description Handle updating those member properties that can be exposed publicly, i.e. not settings
// @Accept multipart/form-data json
// @Param member_name path string true "The nickname of the member being updated"
// @Param Authorization header string true "The JWT token"
// @Param X-CSRF-Token header string true "CSRF token"
// @Param profile_pic_id query int64 false "ID of the picture that is returned after making a request to /api/upload/image"
// @Success 200 {object} h.ResponseHTTP{}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /update/{member_name} [patch]
func (mc *Controller) Update(c *fiber.Ctx) (err error) {
	mc.log.Info().Msg("Update called")
	var member *member.Member
	ct := c.Request().Header.Peek("Content-Type")
	if strings.Contains(string(ct), "multipart/form-data") {
		member, err = parseFormValues(c)
		if err != nil {
			return h.BadRequest(mc.log, c, "Invalid update request", err.Error(), err)
		}
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

// UpdatePrefs handles the updating of user preferences
// @Summary Update member preferences
// @Description Handle updating private member preferences
// @Tags accounts,updating,settings
// @Accept json multipart/form-data
// @Param member_name path string true "The nickname of the member being updated"
// @Param Authorization header string true "The JWT token"
// @Param X-CSRF-Token header string true "CSRF token"
// @Param locale formData string false "The ISO 639-1 locale to use"
// @Param rating_scale_lower formData int16 false "The lower bound of the rating scale" minimum(0) maximum(100)
// @Param rating_scale_upper formData int16 false "The upper bound of the rating scale" minimum(2) maximum(100)
// @Param message_autohide_words formData []string false "A comma-separated list of words to autohide in messages"
// @Param muted_instances formData []string false "A comma-separated list of instance domains to mute"
// @Param auto_accept_follow formData bool false "Whether to automatically accept follow requests"
// @Param locally_searchable formData bool false "Whether to allow local searches"
// @Param federated_searchable formData bool false "Whether to allow federated searches"
// @Param robots_searchable formData bool false "Whether to allow robots to index the profile"
// @Success 200 {object} h.ResponseHTTP{}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /update/{member_name}/preferences [patch]
func (mc *Controller) UpdatePrefs(c *fiber.Ctx) error {
	mc.log.Info().Msg("UpdatePrefs called")
	var prefs *member.Preferences
	var err error
	ct := c.Request().Header.Peek("Content-Type")
	if strings.Contains(string(ct), "multipart/form-data") {
		prefs, err = parseFormPrefs(c)
		if err != nil {
			mc.log.Error().Err(err).Msgf("Error parsing form values: %v", err)
			return h.Res(c, fiber.StatusBadRequest, "Error parsing form values")
		}
		mc.log.Debug().Msgf("prefs: %+v", prefs)
	} else {
		err = c.BodyParser(&prefs)
		if err != nil {
			mc.log.Error().Err(err).Msgf("Error parsing request body: %v", err)
			return h.Res(c, fiber.StatusBadRequest, "Error parsing request body")
		}
	}
	return h.Res(c, fiber.StatusNotImplemented, "Preferences updating not implemented yet")
}

func parseFormPrefs(c *fiber.Ctx) (p *member.Preferences, err error) {
	var tag language.Tag
	if c.FormValue("locale") != "" {
		tag, err = language.Parse(c.FormValue("locale"))
		if err != nil {
			return nil, h.Res(c, fiber.StatusBadRequest, "Invalid locale")
		}
	} else {
		tag = language.English
		// TODO: get the current locale from the database
		// maybe when the data is sent to the client, automatically include non-modified fields
		// in the request
	}
	lower, err := strconv.ParseInt(c.FormValue("rating_scale_lower", "1"), 10, 16)
	if err != nil {
		return nil, h.Res(c, fiber.StatusBadRequest, "Invalid rating scale lower bound")
	}
	upper, err := strconv.ParseInt(c.FormValue("rating_scale_upper", "10"), 10, 16)
	if err != nil {
		return nil, h.Res(c, fiber.StatusBadRequest, "Invalid rating scale upper bound")
	}

	autoHideWords := strings.Split(c.FormValue("message_autohide_words"), ",")
	muteInstances := strings.Split(c.FormValue("muted_instances"), ",")

	return &member.Preferences{
		UX: member.UXPreferences{
			Locale:           tag,
			RatingScaleLower: int16(lower),
			RatingScaleUpper: int16(upper),
		},
		PrivacySecurity: member.PrivacySecurityPreferences{
			MessageAutohideWords: pq.StringArray(autoHideWords),
			MutedInstances:       pq.StringArray(muteInstances),
			AutoAcceptFollow:     c.FormValue("auto_accept_follow", "true") == "true",
			LocallySearchable:    c.FormValue("locally_searchable", "true") == "true",
			FederatedSearchable:  c.FormValue("federated_searchable", "true") == "true",
			RobotsSearchable:     c.FormValue("robots_searchable", "false") == "true",
		},
	}, nil
}

// parseFormValues parses the form values from the request body for which member struct fields are present
func parseFormValues(c *fiber.Ctx) (m *member.Member, err error) {
	var mm member.Member
	displayName := c.FormValue("display_name")
	var dn, bioValue sql.NullString
	dn = lo.Ternary(displayName != "", sql.NullString{String: displayName, Valid: true}, sql.NullString{Valid: false})
	if dn.Valid {
		mm.DisplayName = dn
	}
	email := c.FormValue("email")
	if email != "" {
		_, err = mail.ParseAddress(email)
		if err != nil {
			return nil, h.Res(c, fiber.StatusBadRequest, "Invalid email")
		}
		mm.Email = email
	}
	bio := c.FormValue("bio")
	bioValue = lo.Ternary(bio != "", sql.NullString{String: bio, Valid: true}, sql.NullString{Valid: false})
	if bioValue.Valid {
		mm.Bio = bioValue
	}

	visibility := c.FormValue("visibility")
	if visibility != "" {
		visibility = strings.ToLower(visibility)
		if visibility != "public" && visibility != "followers_only" && visibility != "private" && visibility != "local" {
			return nil, fmt.Errorf(
				"invalid visibility: %v (was expecting public, followers_only, local, or private)",
				visibility)
		}
		mm.Visibility = visibility
	}
	customFieldsData := c.FormValue("custom_fields")
	if customFieldsData != "" {
		var customFields pgtype.JSONB
		err := customFields.UnmarshalJSON([]byte(customFieldsData))
		if err != nil {
			return nil, fmt.Errorf("error parsing custom fields '%s': %v", customFieldsData, err)
		}
		mm.CustomFields = customFields
	}

	return &mm, nil
}
