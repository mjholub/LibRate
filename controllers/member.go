package controllers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models"
)

// MemberController allows for the retrieval of user information
type MemberController struct {
	storage models.MemberStorage
}

func NewMemberController(storage models.MemberStorage) *MemberController {
	return &MemberController{storage: storage}
}

// GetMember retrieves user information based on the user ID
func (mc *MemberController) GetMember(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	member, err := mc.storage.Read(ctx, c.Params("id"), c.Params("id"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}

	return c.JSON(member)
}

// UpdateMember handles the updating of user information
func (mc *MemberController) UpdateMember(c *fiber.Ctx) error {
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
	err = mc.storage.Update(ctx, member)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to update member info")
	}

	return c.JSON(fiber.Map{
		"message": "Member updated successfully",
	})
}

// DeleteMember handles the deletion of a user
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
