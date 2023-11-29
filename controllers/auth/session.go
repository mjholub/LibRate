package auth

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

// TODO: add saving logic (needed for bans etc)
type SessionData struct {
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	DeviceUUID uuid.UUID `json:"device_uuid"`
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
}

func (a *Service) createSession(c *fiber.Ctx, member *member.Member) error {
	var deviceHash string
	if c.Cookies("device_id") == "" {
		deviceID, err := a.identifyDevice()
		if err != nil {
			a.log.Error().Err(err).Msgf("Failed to create session: %s", err.Error())
			return h.Res(c, http.StatusInternalServerError, "Failed to create session")
		}
		deviceHash = deviceID.String()
		c.Cookie(&fiber.Cookie{
			Domain:      a.conf.Fiber.Domain,
			SessionOnly: true,
			Expires:     time.Now().Add(time.Hour * 24 * 90),
			SameSite:    "Lax",
			Name:        "device_id",
			Value:       deviceHash,
			HTTPOnly:    false,
		})
	} else {
		deviceHash = c.Cookies("device_id")
	}

	sess, err := a.sess.Get(c)
	if err != nil {
		a.log.Error().Err(err).Msgf("Failed to create session: %s", err.Error())
		return h.Res(c, http.StatusInternalServerError, "Failed to create session")
	}

	// TODO: add custom session expiry

	sess.Set("memberName", member.MemberName)
	sess.Set("session_id", sess.ID())
	sess.Set("device_id", deviceHash)
	token, err := a.createToken()
	if err != nil {
		a.log.Error().Err(err).Msgf("Failed to create session: %s", err.Error())
		return h.Res(c, http.StatusInternalServerError, "Failed to create session")
	}
	if err = a.sess.Storage.Set("token", []byte(token), time.Hour*24*90); err != nil {
		a.log.Error().Err(err).Msgf("Failed to create session: %s", err.Error())
		return h.Res(c, http.StatusInternalServerError, "Failed to create session")
	}

	if err = sess.Save(); err != nil {
		a.log.Error().Err(err).Msgf("Failed to create session: %s", err.Error())
		return h.Res(c, http.StatusInternalServerError, "Failed to create session")
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":    "Logged in successfully",
		"memberName": member.MemberName,
		"token":      token,
	})
}

func (a *Service) GetAuthStatus(c *fiber.Ctx) error {
	a.log.Debug().Msg("GetAuthStatus request")
	token := c.Request().Header.Peek("Authorization")
	// TODO: remove
	if token == nil {
		a.log.Warn().Msg("No JWT token found")
	}

	csrfToken := c.Cookies("csrf_")
	if csrfToken == "" {
		a.log.Warn().Msgf("No CSRF token found for request on %s on %s", c.OriginalURL(), c.IP())
		return h.Res(c, http.StatusForbidden, "Forbidden")
	}

	sess, err := a.sess.Get(c)
	if err != nil {
		return h.Res(c, http.StatusInternalServerError, "Failed to get session")
	}
	if c.Cookies("session_id") == "" {
		a.log.Warn().Msg("No session cookie found")
		return h.Res(c, http.StatusUnauthorized, "Not logged in")
	}

	sessionID := sess.ID()
	sessionFallback := sess.Get(c.Cookies("session_id"))
	if sessionID == "" && sessionFallback == nil {
		a.log.Warn().Msg("No session ID found")
		a.log.Debug().Msgf("session keys: %+v", sess.Keys())
		return h.Res(c, http.StatusUnauthorized, "Not logged in")
	}

	if sessionID != c.Cookies("session_id") {
		a.log.Warn().Msg("Session ID mismatch")
		a.log.Debug().Msgf("Session ID: %s", sessionID)
		a.log.Debug().Msgf("Cookie session ID: %s", c.Cookies("session_id"))
		return h.Res(c, http.StatusUnauthorized, "Not logged in")
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":         "Logged in",
		"isAuthenticated": true,
		"memberName":      sess.Get("memberName"),
	})
}

// TODO: create corresponding database modifications so that we can tie a device to a member
func (a *Service) identifyDevice() (uuid.UUID, error) {
	deviceID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	return deviceID, nil
}

// GetSessionTimeoutPrefs returns the session timeout preferences
// This setting is not synced across devices.
// The way it works is:
// 1. Send a request to the database, where a JOIN is performed on the current device's identifier and the member's ID
// 2. If the device is found, return the timeout preference
func (a *Service) GetSessionTimeoutPrefs(c *fiber.Ctx) error {
	memberStr := c.Params("memberUUID")
	if memberStr == "" {
		return h.Res(c, http.StatusBadRequest, "Member UUID missing or member not found")
	}

	deviceStr := c.Cookies("device_id")
	if deviceStr == "" {
		return h.Res(c, http.StatusBadRequest, "Device ID not found")
	}

	// sanitize the received parameters, don't trust random strings
	memberID, err := strconv.Atoi(memberStr)
	if err != nil {
		return h.Res(c, http.StatusBadRequest, "Invalid member UUID")
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
