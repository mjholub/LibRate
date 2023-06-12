package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/models"

	"github.com/gofiber/fiber/v2"
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

func getMemberData(lookupField, lookupTarget string) (*models.Member, error) {
	conf := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	dbConn, err := db.Connect(&conf)
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()

	ms := models.NewMemberStorage(dbConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	member, err := ms.Read(ctx, lookupField, lookupTarget)
	if err != nil {
		return nil, err
	}

	return member, nil
}

func validatePassword(member *models.Member, password string) error {
	if !checkArgonPassword(member.PassHash, password) {
		return errors.New("invalid email or password")
	}

	return nil
}

func Login(c *fiber.Ctx) error {
	input, err := parseInput("login", c)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid login request",
		})
	}
	var lookupField, lookupTarget string

	validatedInput, err := input.Validate()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid login request",
		})
	}
	switch {
	case validatedInput.Email == "" && validatedInput.MemberName != "":
		lookupTarget = validatedInput.MemberName
		lookupField = "nick"
	case validatedInput.Email != "" && validatedInput.MemberName == "":
		lookupTarget = validatedInput.Email
		lookupField = "email"
	default:
		lookupTarget = validatedInput.MemberName
		lookupField = "nick"
	}

	member, err := getMemberData(lookupField, lookupTarget)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to connect to database: %s" + err.Error(),
		})
	}

	err = validatePassword(member, validatedInput.Password)

	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Logged in successfully",
	})
}
