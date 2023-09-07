package auth

import (
	"errors"
	"net/http"

	"codeberg.org/mjh/LibRate/models"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

// 1. Parse the input
// 2. Validate the input (check for empty fields, valid email, etc.)
// 3. Pass the email to the database, get the password hash for the email or nickname
// 4. Compare the password hash with the password hash from the database
func (a *AuthService) Login(c *fiber.Ctx) error {
	a.log.Debug().Msg("Login request")
	input, err := parseInput("login", c)
	if err != nil {
		return h.Res(c, http.StatusBadRequest, "Invalid login request")
	}
	if input == nil {
		return h.Res(c, http.StatusInternalServerError, "Cannot parse input")
	}
	a.log.Debug().Msg("Parsed input")

	validatedInput, err := input.Validate()
	if err != nil {
		return h.Res(c, http.StatusBadRequest, "Invalid login request")
	}
	a.log.Debug().Msg("Validated input")

	err = a.validatePassword(
		validatedInput.Email,
		validatedInput.MemberName,
		validatedInput.Password,
	)

	member := models.Member{
		ID:         0,
		Email:      validatedInput.Email,
		MemberName: validatedInput.MemberName,
		PassHash:   validatedInput.Password,
	}

	switch {
	case validatedInput.Email != "" && err == nil:
		memberID, err := a.ms.GetID(c.Context(), validatedInput.Email)
		if err != nil {
			return h.Res(c, http.StatusInternalServerError, "Failed to validate credentials")
		}
		member.ID = memberID
		return a.createSession(c, &member)
	case validatedInput.MemberName != "" && err == nil:
		memberID, err := a.ms.GetID(c.Context(), validatedInput.MemberName)
		if err != nil {
			return h.Res(c, http.StatusInternalServerError, "Failed to validate credentials")
		}
		member.ID = memberID
		return a.createSession(c, &member)
	case err != nil && a.conf.LibrateEnv == "dev":
		return h.Res(c, http.StatusUnauthorized, "Invalid credentials: "+err.Error())
	case err != nil:
		return h.Res(c, http.StatusUnauthorized, "Invalid credentials")
	default:
		return h.Res(c, http.StatusInternalServerError, "Internal server error")
	}
}

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
	passhash, err := a.ms.GetPassHash(email, login)
	if err != nil {
		return err
	}
	if !checkArgonPassword(password, passhash) {
		return errors.New("invalid email, username or password")
	}

	return nil
}

// TODO: move to a dedicated file?
func (a *AuthService) createSession(c *fiber.Ctx, member *models.Member) error {
	token, err := a.ms.CreateSession(c.Context(), *member)
	if err != nil {
		return h.Res(c, http.StatusInternalServerError, "Internal server error")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    token,
		HTTPOnly: true,
	})

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Logged in successfully",
		"token":   token,
	})
}
