package auth

import (
	"errors"
	"net/http"

	"codeberg.org/mjh/LibRate/models"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func (l LoginInput) Validate() (*models.MemberInput, error) {
	if l.Email == "" && l.MemberName == "" {
		return nil, errors.New("email or nickname required")
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

func (a *AuthService) validatePassword(email, login, password string) error {
	ms := models.NewMemberStorage(a.db)

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
func (a *AuthService) Login(c *fiber.Ctx) error {
	input, err := parseInput("login", c)
	if err != nil {
		return h.Res(c, http.StatusBadRequest, "Invalid login request")
	}

	validatedInput, err := input.Validate()
	if err != nil {
		return h.Res(c, http.StatusBadRequest, "Invalid login request")
	}

	err = a.validatePassword(
		validatedInput.Email,
		validatedInput.MemberName,
		validatedInput.Password)
	if err != nil && a.conf.LibrateEnv == "dev" {
		return h.Res(c, http.StatusUnauthorized, "Invalid credentials: "+err.Error())
	} else if err != nil {
		return h.Res(c, http.StatusUnauthorized, "Invalid credentials")
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Logged in successfully",
	})
}
