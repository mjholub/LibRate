package auth

import (
	"errors"
	"fmt"
	"net/http"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func (l LoginInput) Validate() (*models.MemberInput, error) {
	if l.Email == "" && l.MemberName == "" {
		return nil, errors.New("email or membername required")
	}

	if l.Password == "" {
		return nil, errors.New("password required")
	}

	return &models.MemberInput{
		Email:      l.Email,
		MemberName: l.MemberName,
		Password:   l.Password,
	}, nil
}

func validatePassword(dbConn *sqlx.DB, email, login, password string) error {
	ms := models.NewMemberStorage(dbConn)

	passhash, err := ms.GetPassHash(email, login)
	if err != nil {
		return err
	}
	if !checkArgonPassword(password, passhash) {
		return errors.New("invalid email, username or password")
	}

	return nil
}

// TODO: verify if the database connection can be passed in as a parameter
func Login(c *fiber.Ctx) error {
	conf := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	dbConn, err := db.Connect(&conf)
	defer dbConn.Close()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to connect to database: %s" + err.Error(),
		})
	}
	input, err := parseInput("login", c)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Invalid login request")
	}

	validatedInput, err := input.Validate()
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Invalid login request")
	}

	err = validatePassword(dbConn, validatedInput.Email, validatedInput.MemberName, validatedInput.Password)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "Invalid credentials: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Logged in successfully",
	})
}

func errorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"message": message,
	})
}
