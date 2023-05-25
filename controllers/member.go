package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models"
)

// GetMember retrieves user information based on the user ID
func GetMember(c *fiber.Ctx) error {
	conf := cfg.LoadDgraph()
	ms, err := models.NewMemberStorage(*conf)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize member storage",
		})
	}
	member, err := ms.Load(context.TODO(), c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Member not found",
		})
	}

	return c.JSON(member)
}

// CreateMember handles the creation of a new user
func CreateMember(c *fiber.Ctx) error {
	var input models.MemberInput
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	m := models.Member{
		MemberName: input.MemberName,
		Email:      input.Email,
	}

	conf := cfg.LoadDgraph()
	ms, err := models.NewMemberStorage(*conf)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize member storage",
		})
	}

	err = ms.Save(context.TODO(), &m)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	return c.JSON(m)
}

func UpdateMember(c *fiber.Ctx) error {
	var input models.MemberInput
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	conf := cfg.LoadDgraph()
	ms, err := models.NewMemberStorage(*conf)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to database",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	member, err := ms.Load(ctx, input.Email)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Member not found",
		})
	}
	err = ms.Update(ctx, member)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Member updated successfully",
	})
}

// DeleteMember handles the deletion of a user
func DeleteMember(c *fiber.Ctx) error {
	var input models.MemberInput
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	conf := cfg.LoadDgraph()
	ms, err := models.NewMemberStorage(*conf)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to database",
		})
	}
	member, err := ms.Load(context.Background(), input.Email)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": "Failed to load member",
		})
	}
	err = ms.Delete(context.Background(), member)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Member deleted successfully",
	})
}
