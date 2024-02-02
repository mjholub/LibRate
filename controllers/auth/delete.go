package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// @Summary Delete account
// @Description Delete the account of the currently logged in user
// @Tags auth,accounts,deleting,settings
// @Accept json
// @Param password body string true "The password"
// @Param confirmation body string true "Confirmation of the password"
// @Param X-CSRF-Token header string true "CSRF protection token"
// @Param Authorization header string true "JWT token"
// @Router /authenticate/delete-account [post]
func (a *Service) DeleteAccount(c *fiber.Ctx) error {
	a.log.Debug().Msg("Delete account request")
	memberName := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["member_name"].(string)
	if memberName == "" {
		return h.Res(c, fiber.StatusUnauthorized, "Not logged in")
	}

	passHash, err := a.ms.GetPassHash("", memberName)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to retrieve password hash")
	}
	oldPass := c.Params("password")

	if !checkArgonPassword(oldPass, passHash) {
		return h.Res(c, fiber.StatusUnauthorized, "Invalid password")
	}

	confirm := c.Params("confirmation")
	if oldPass != confirm {
		return h.Res(c, fiber.StatusBadRequest, "Passwords do not match")
	}

	err = a.ms.Delete(c.Context(), memberName)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to delete account")
	}
	return nil
}
