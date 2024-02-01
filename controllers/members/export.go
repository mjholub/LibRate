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
func (mc *MemberController) Export(c *fiber.Ctx) error {
	memberName := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["member_name"].(string)
	if memberName == "" {
		return h.BadRequest(mc.log, c, "missing name in JWT token", "export request initialized without a token from "+c.IP(), nil)
	}
	format := c.Params("format")
	availableFormats := []string{"json", "csv", "sql"}
	if !lo.Contains(availableFormats, format) {
		return h.BadRequest(
			mc.log,
			c,
			"invalid format",
			fmt.Sprintf("Member %s (IP: %s) tried to initialize export with invalid format '%s'", memberName, c.IP(), format), nil)
	}
	mc.log.Info().Msgf("%s initialized a data export request using %s format", memberName, format)

	switch format {
	case "json":
		return mc.exportJSON(c)
	case "csv":
		return mc.exportCSV(c)
	case "sql":
		return mc.exportSQL(c)
	default:
		// should never happen due to the check above, this is just to satisfy the compiler
		return c.SendStatus(fiber.StatusInternalServerError)
	}
}

func (mc *MemberController) exportJSON(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (mc *MemberController) exportCSV(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (mc *MemberController) exportSQL(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}
