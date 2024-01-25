package members

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

// @Summary Ban user
// @Description issues a ban to a user with the given UUID
// @Tags members,accounts,administration
// @Accept json
// @Produce json
// @Param uuid path string true "UUID of the member to ban"
// @Param input body BanInput true "Ban details"
// @Param X-CSRF-Token header string true "X-CSRF-Token header"
// @Success 200 {object} h.ResponseHTTP{}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/{uuid}/ban [post]
func (mc *MemberController) Ban(c *fiber.Ctx) error {
	requester := c.Locals("jwtToken").(*jwt.Token)
	claims := requester.Claims.(jwt.MapClaims)
	name := claims["member_name"].(string)
	if !mc.storage.HasRole(c.Context(), name, "mod", true) || !mc.storage.HasRole(c.Context(), name, "admin", true) {
		mc.log.Warn().Msgf("Member %s tried to ban a member", name)
		return h.Res(c, 403, "Forbidden")
	}
	banInput := member.BanInput{}

	if err := c.BodyParser(&banInput); err != nil {
		mc.log.Error().Err(err).Msg("Failed to parse ban input")
		return h.Res(c, 400, "Invalid input")
	}
	id, err := uuid.FromString(c.Params("uuid"))
	if err != nil {
		mc.log.Error().Err(err).Msg("Failed to parse UUID")
		return h.Res(c, 400, "Invalid UUID")
	}
	m := member.Member{
		UUID: id,
	}

	err = mc.storage.Ban(c.Context(), &m, &banInput)
	if err != nil {
		mc.log.Error().Err(err).Msg("Failed to ban member")
		return h.Res(c, 500, "Failed to ban member")
	}

	return h.Res(c, 200, "OK")
}

// @Summary Unban user
// @Description removes a ban from a user with the given UUID
// @Tags members,accounts,administration
// @Accept json
// @Produce json
// @Param uuid path string true "UUID of the member to unban"
// @Param X-CSRF-Token header string true "X-CSRF-Token header"
// @Success 200 {object} h.ResponseHTTP{}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/{uuid}/ban [delete]
func (mc *MemberController) Unban(c *fiber.Ctx) error {
	requester := c.Locals("jwtToken").(*jwt.Token)
	claims := requester.Claims.(jwt.MapClaims)
	name := claims["member_name"].(string)
	if !mc.storage.HasRole(c.Context(), name, "moderator") {
		mc.log.Warn().Msgf("Member %s tried to unban a member", name)
		return h.Res(c, 403, "Forbidden")
	}
	id, err := uuid.FromString(c.Params("uuid"))
	if err != nil {
		mc.log.Error().Err(err).Msg("Failed to parse UUID")
		return h.Res(c, 400, "Invalid UUID")
	}
	m := member.Member{
		UUID: id,
	}

	err = mc.storage.Unban(c.Context(), &m)
	if err != nil {
		mc.log.Error().Err(err).Msg("Failed to unban member")
		return h.Res(c, 500, "Failed to unban member")
	}

	return h.Res(c, 200, "OK")
}
