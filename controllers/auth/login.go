package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"codeberg.org/mjh/LibRate/models/member"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// 1. Parse the input
// 2. Validate the input (check for empty fields, valid email, etc.)
// 3. Pass the email to the database, get the password hash for the email or nickname
// 4. Compare the password hash with the password hash from the database
func (a *Service) Login(c *fiber.Ctx) error {
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

	member := member.Member{
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

func (a *Service) createSession(c *fiber.Ctx, member *member.Member) error {
	token, err := a.createToken()
	if err != nil {
		return h.Res(c, http.StatusInternalServerError, "Failed to create session")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    token,
		HTTPOnly: true,
	})

	if member.ID == 0 {
		memberID, err := a.ms.GetID(c.Context(), member.Email)
		if err != nil {
			return h.Res(c, http.StatusInternalServerError, "Failed to validate credentials")
		}
		member.ID = memberID
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":   "Logged in successfully",
		"token":     token,
		"member_id": member.ID,
	})
}

func (a *Service) createToken() (string, error) {
	token, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return token.String(), nil
}

// GetSessionTimeoutPrefs returns the session timeout preferences
// This setting is not synced across devices.
// The way it works is:
// 1. Send a request to the database, where a JOIN is performed on the current device's identifier and the member's ID
// 2. If the device is found, return the timeout preference
func (a *Service) GetSessionTimeoutPrefs(c *fiber.Ctx) error {
	memberStr := c.Params("member_id")
	if memberStr == "" {
		return h.Res(c, http.StatusBadRequest, "Member ID missing or member not found")
	}

	deviceStr := c.Cookies("device_id")
	if deviceStr == "" {
		return h.Res(c, http.StatusBadRequest, "Device ID not found")
	}

	// sanitize the received parameters, don't trust random strings
	memberID, err := strconv.Atoi(memberStr)
	if err != nil {
		return h.Res(c, http.StatusBadRequest, "Invalid member ID")
	}

	deviceID, err := uuid.FromString(deviceStr)
	if err != nil {
		return h.Res(c, http.StatusBadRequest, "Invalid device ID")
	}

	timeout, err := a.ms.GetSessionTimeout(c.Context(), memberID, deviceID)
	if err != nil {
		return h.Res(c, http.StatusInternalServerError, "Failed to get session timeout preferences")
	}

	return c.Status(http.StatusOK).SendString(strconv.Itoa(timeout))
}

// isKnownDevice queries the database to check if the device has been saved.
// Since we use uuids to assign device IDs, the member ID is redundant
func (a *Service) isKnownDevice(c *fiber.Ctx) (bool, error) {
	deviceStr := c.Cookies("device_id")
	if deviceStr == "" {
		return false, nil
	}

	deviceID, err := uuid.FromString(deviceStr)
	if err != nil {
		return false, h.Res(c, http.StatusBadRequest, "Invalid device ID")
	}

	// query the db
	err = a.ms.LookupDevice(c.Context(), deviceID)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
