package members

import (
	"context"
	"encoding/json"
	"time"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models"
	"github.com/gofiber/fiber/v2"
)

// DeleteMember handles the deletion of an user
func (mc *MemberController) DeleteMember(c *fiber.Ctx) error {
	var input models.MemberInput
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
	err = mc.storage.Delete(ctx, member)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to delete member")
	}

	return c.JSON(fiber.Map{
		"message": "Member deleted successfully",
	})
}
