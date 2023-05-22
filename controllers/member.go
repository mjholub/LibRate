package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"codeberg.org/mjh/LibRate/models"
)

// GetMember retrieves user information based on the user ID
func GetMember(c *fiber.Ctx) error {
	ms := models.NewMemberStorer()
	member, err := ms.Load(context.TODO(), c.Params("id"))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
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

	ms := models.NewMemberStorer()

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

	_, err = models.UpdateMember(input)
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
	// TODO: implement

	ms := models.NewMemberStorer()
	err := ms.Delete(context.Background(), c.Params("id"))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Member deleted successfully",
	})
}
