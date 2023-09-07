package controllers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models"
)

// MemberController allows for the retrieval of user information
type (
	// IMemberController is the interface for the member controller
	// It defines the methods that the member controller must implement
	// This is useful for mocking the member controller in unit tests
	IMemberController interface {
		GetMember(c *fiber.Ctx) error
		UpdateMember(c *fiber.Ctx) error
		DeleteMember(c *fiber.Ctx) error
	}

	// MemberController is the controller for member endpoints
	MemberController struct {
		storage models.MemberStorage
		log     *zerolog.Logger
	}
)

func NewMemberController(storage models.MemberStorage, logger *zerolog.Logger) *MemberController {
	return &MemberController{storage: storage, log: logger}
}

// GetMember retrieves user information based on the user ID
func (mc *MemberController) GetMember(c *fiber.Ctx) error {
	mc.log.Info().Msg("GetMember called")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mc.log.Debug().Msgf("ID: %s", c.Params("id"))
	member, err := mc.storage.Read(ctx, "id", c.Params("id"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	mc.log.Info().Msgf("Member: %+v", member)

	return h.ResData(c, fiber.StatusOK, "success", member)
}

func (mc *MemberController) GetMemberByNick(c *fiber.Ctx) error {
	mc.log.Info().Msg("GetMemberByNick called")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mc.log.Debug().Msgf("Nick: %s", c.Params("nick"))
	member, err := mc.storage.Read(ctx, "nick", c.Params("nick"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	mc.log.Info().Msgf("Member: %+v", member)

	return h.ResData(c, fiber.StatusOK, "success", member)
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
