package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"librerym/models"
)

// GetUser retrieves user information based on the user ID
func GetUser(c *fiber.Ctx) error {
	userID, _ := strconv.Atoi(c.Params("id"))

	user, err := models.GetUserByID(userID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

// CreateUser handles the creation of a new user
func CreateUser(c *fiber.Ctx) error {
	var input models.UserInput
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	user := models.User{
		Username: input.Username,
		Email:    input.Email,
	}

	err = models.SaveUser(&user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	return c.JSON(user)
}

func UpdateUser(c *fiber.Ctx) error {
	userID, _ := strconv.Atoi(c.Params("id"))

	var input models.UserInput
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	err = models.UpdateUser(userID, &input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User updated successfully",
	})
}

// DeleteUser handles the deletion of a user
func DeleteUser(c *fiber.Ctx) error {
	userID, _ := strconv.Atoi(c.Params("id"))

	err := models.DeleteUser(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}
