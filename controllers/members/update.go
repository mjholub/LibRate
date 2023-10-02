package members

import (
	"context"
	"encoding/json"
	"time"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
	"github.com/gofiber/fiber/v2"
)

// UpdateMember handles the updating of user information
func (mc *MemberController) UpdateMember(c *fiber.Ctx) error {
	var input member.Input
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid input")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	member, err := mc.storage.Read(ctx, "email", input.Email)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	err = mc.storage.Update(ctx, member)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to update member info")
	}

	return c.JSON(fiber.Map{
		"message": "Member updated successfully",
	})
}
