package auth

import (
	"errors"
	"net/http"

	"codeberg.org/mjh/LibRate/models/member"

	"github.com/gofiber/fiber/v2"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// 1. Parse the input
// 2. Validate the input (check for empty fields, valid email, etc.)
// 3. Pass the email to the database, get the password hash for the email or nickname
// 4. Compare the password hash with the password hash from the database
func (a *Service) Login(c *fiber.Ctx) error {
	a.log.Debug().Msg("Login request")
	input, err := parseLoginInput(c)
	if err != nil {
		a.log.Debug().Msgf("Failed to parse input: %s", err.Error())
		return h.Res(c, http.StatusBadRequest, "Invalid login request")
	}
	if input == nil {
		return h.Res(c, http.StatusInternalServerError, "Cannot parse input")
	}
	a.log.Debug().Msgf("Parsed input")

	validatedInput, err := input.Validate()
	if err != nil {
		return h.Res(c, http.StatusBadRequest, "Invalid login request")
	}
	if validatedInput == nil {
		return h.Res(c, http.StatusInternalServerError, "Cannot parse input")
	}
	a.log.Debug().Msg("Validated input")

	err = a.validatePassword(
		validatedInput.Email,
		validatedInput.MemberName,
		validatedInput.Password,
	)
	if err != nil {
		if a.conf.LibrateEnv == "development" {
			a.log.Debug().Msgf("Failed to validate password: %s", err.Error())
			return h.Res(c, http.StatusUnauthorized, "Invalid credentials")
		}
		return h.Res(c, http.StatusUnauthorized, "Invalid credentials")
	}

	member := member.Member{
		Email:      validatedInput.Email,
		MemberName: validatedInput.MemberName,
		PassHash:   validatedInput.Password,
	}

	return a.createSession(c, input.RememberMe, &member)
}

func (l LoginInput) Validate() (*member.Input, error) {
	if l.Email == "" && l.MemberName == "" {
		return nil, errors.New("email or nickname required")
	}

	if l.Password == "" {
		return nil, errors.New("password required")
	}

	return &member.Input{
		Email:      l.Email,
		MemberName: l.MemberName,
		Password:   l.Password,
	}, nil
}

func (a *Service) validatePassword(email, login, password string) error {
	passhash, err := a.ms.GetPassHash(email, login)
	if err != nil {
		return err
	}
	if !checkArgonPassword(password, passhash) {
		return errors.New("invalid email, username or password")
	}

	return nil
}
