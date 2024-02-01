package members

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// @Summary Export all of the member's data
// @Description Exports the data of a member, including profile information as well as other related data such as reviews
// @Tags accounts,members,metadata
// @Accept json
// @Produce json text/csv
// @Param Authorization header string true "JWT access token"
// @Param format path string true "Export format" Enums(json, csv)
// @Router /members/export/{format} [get]
func (mc *MemberController) Export(c *fiber.Ctx) error {
	memberName := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["member_name"].(string)
	if memberName == "" {
		return h.BadRequest(mc.log, c, "missing name in JWT token", "export request initialized without a token from "+c.IP(), nil)
	}
	format := c.Params("format")
	availableFormats := []string{"json", "csv"}
	if !lo.Contains(availableFormats, format) {
		return h.BadRequest(
			mc.log,
			c,
			"invalid format",
			fmt.Sprintf("Member %s (IP: %s) tried to initialize export with invalid format '%s'", memberName, c.IP(), format), nil)
	}
	mc.log.Info().Msgf("%s initialized a data export request using %s format", memberName, format)

	data, err := mc.storage.Export(c.Context(), memberName, format)
	if err != nil {
		return h.InternalError(mc.log, c, fmt.Sprintf("failed to export data for %s using %s", memberName, format), err)
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", memberName, format))
	switch format {
	case "json":
		c.Set("Content-Type", "application/json")
	case "csv":
		c.Set("Content-Type", "text/csv")
	}
	return c.Send(data)
}
