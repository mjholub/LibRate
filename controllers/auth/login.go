package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"codeberg.org/mjh/LibRate/models/member"

	"github.com/gofiber/fiber/v2"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// 1. Parse the input
// 2. Validate the input (check for empty fields, valid email, etc.)
// 3. Pass the email to the database, get the password hash for the email or nickname
// 4. Compare the password hash with the password hash from the database
// @Summary Login to the application
// @Description Create a session for the user
// @Tags auth,accounts
// @Accept multipart/form-data
// @Produce json
// @Param membername query string false "Member name. Request must include either membername or email"
// @Param email query string false "Email address"
// @Param session_time query int false "Session time in minutes" default(30) minimum(1) maximum(2147483647)
// @Param password query string true "Password"
// @Param Referrer-Policy header string false "Referrer-Policy header" "no-referrer-when-downgrade"
// @Param X-CSRF-Token header string true "X-CSRF-Token header"
// @Success 200 {object} h.ResponseHTTP{data=SessionResponse}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /authenticate/login [post]
func (a *Service) Login(c *fiber.Ctx) error {
	a.log.Debug().Msg("Login request")
	maybeInput, err := parseLoginInput(c, a.log)
	if err != nil {
		a.log.Debug().Msgf("Failed to parse input: %s", err.Error())
		return h.Res(c, http.StatusBadRequest, "Invalid login request")
	}
	input, isSome := maybeInput.Get()
	if !isSome {
		return h.Res(c, http.StatusInternalServerError, "Error parsing input")
	}
	a.log.Debug().Msgf("Parsed input: %+v", input)

	err = a.validatePassword(
		c.Context(),
		input.Email,
		input.MemberName,
		input.Password,
	)
	if err != nil {
		if a.conf.LibrateEnv == "development" {
			a.log.Debug().Msgf("Failed to validate password: %s", err.Error())
			return h.Res(c, http.StatusUnauthorized, "Invalid credentials")
		}
		return h.Res(c, http.StatusUnauthorized, "Invalid credentials")
	}
	a.log.Debug().Msg("Validated password")

	memberData := member.Member{
		Email:      input.Email,
		MemberName: input.MemberName,
		Webfinger:  fmt.Sprintf("%s@%s", input.MemberName, a.conf.Fiber.Domain),
		PassHash:   input.Password,
	}

	return a.createSession(c, input.SessionTime, &memberData)
}

func (a *Service) validatePassword(ctx context.Context, email, login, password string) error {
	passhash, err := a.ms.GetPassHash(ctx, email, login)
	if err != nil {
		return err
	}
	if !checkArgonPassword(password, passhash) {
		return errors.New("invalid email, username or password")
	}

	return nil
}
