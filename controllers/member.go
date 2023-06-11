package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/models"
)

// GetMember retrieves user information based on the user ID
func GetMember(c *fiber.Ctx) error {
	conf := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	// TODO: wrap handling db errors into monads as well
	dbConn, err := db.Connect(&conf)
	ms := models.NewMemberStorage(dbConn)
	defer dbConn.Close()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize member storage",
		})
	}
	member, err := ms.Read(context.TODO(), c.Params("id"), c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Member not found",
		})
	}

	return c.JSON(member)
}

func UpdateMember(c *fiber.Ctx) error {
	var input models.MemberInput
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	conf := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	// TODO: wrap handling db errors into monads as well
	dbConn, err := db.Connect(&conf)
	ms := models.NewMemberStorage(dbConn)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize member storage",
		})
	}
	defer dbConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	member, err := ms.Read(ctx, "email", input.Email)
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

	conf := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	// TODO: wrap handling db errors into monads as well
	dbConn, err := db.Connect(&conf)
	ms := models.NewMemberStorage(dbConn)
	defer dbConn.Close()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize member storage",
		})
	}
	member, err := ms.Read(context.Background(), "email", input.Email)
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
