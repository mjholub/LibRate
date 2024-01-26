package auth

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofrs/uuid/v5"

	"github.com/golang-jwt/jwt/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

type SessionData struct {
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	DeviceUUID uuid.UUID `json:"device_uuid"`
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
}

type SessionResponse struct {
	Token      string `json:"token" example:"[A-Za-z0-9]{37}.[A-Za-z0-9]{147}.L-[A-Za-z0-9]{24}_[A-Za-z0-9]{25}-zNjCwGMr-[A-Za-z0-9]{27}"`
	MemberName string `json:"membername" example:"lain"`
}

func (a *Service) createSession(c *fiber.Ctx, timeout int32, memberData *member.Member) error {
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
			Expires:     time.Now().Add(time.Minute * time.Duration(timeout)),
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
	if sess == nil {
		a.log.Error().Msg("Failed to create session: session is nil")
		return h.Res(c, http.StatusInternalServerError, "Failed to create session")
	}
	a.log.Debug().Msgf("Creating session with ID: %s", sess.ID())
	var mu sync.Mutex

	// TODO: add lock acquisition timeout
	mu.Lock()
	go sess.Set("member_name", memberData.MemberName)
	go sess.Set("session_id", sess.ID())
	go sess.Set("device_id", deviceHash)
	go sess.Set("ip", c.IP())
	go sess.Set("user_agent", string(c.Request().Header.UserAgent()))
	sessionExpiry := time.Duration(timeout) * time.Minute
	sess.SetExpiry(sessionExpiry)
	mu.Unlock()

	signedToken, err := a.createToken(memberData, &sessionExpiry, sess)
	if err != nil {
		a.log.Err(err)
		return h.Res(c, fiber.StatusInternalServerError, "Failed to prepare session")
	}
	c.Cookie(&fiber.Cookie{
		HTTPOnly: true,
		Name:     "session_id",
		MaxAge:   int(sessionExpiry.Seconds()),
		Domain:   a.conf.Fiber.Domain,
		SameSite: "Lax",
		Value:    sess.ID(),
	},
	)

	if err = sess.Save(); err != nil {
		a.log.Error().Err(err).Msgf("Failed to create session: %s", err.Error())
		return h.Res(c, http.StatusInternalServerError, "Failed to create session")
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":    "Logged in successfully",
		"token":      signedToken,
		"memberName": memberData.MemberName,
	})
}

func (a *Service) createToken(memberData *member.Member, timeout *time.Duration, sess *session.Session) (t string, err error) {
	claims := jwt.MapClaims{
		"member_name": memberData.MemberName,
		"webfinger":   memberData.Webfinger,
		"session_id":  sess.ID(),
		"roles":       []string{"member"},
		"exp":         time.Now().Add(*timeout).Unix(),
	}
	sess.Set("claims_exp", claims["exp"])
	sess.Set("claims_roles", claims["roles"])

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	if token == nil {
		return "", errors.New("failed to create token")
	}

	signedToken, err := token.SignedString([]byte(a.conf.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to create a signed token: %v", err)
	}

	return signedToken, nil
}

func (a *Service) GetAuthStatus(c *fiber.Ctx) error {
	a.log.Debug().Msg("GetAuthStatus request")
	token := c.Request().Header.Peek("Authorization")
	if token == nil {
		a.log.Warn().Msg("No token found")
		return h.Res(c, fiber.StatusUnauthorized, "Not logged in")
	}

	csrfToken := c.Cookies("csrf_")
	if csrfToken == "" {
		a.log.Warn().Msgf("No CSRF token found for request on %s on %s", c.OriginalURL(), c.IP())
		return h.Res(c, http.StatusForbidden, "Forbidden")
	}
	a.log.Trace().Msg("CSRF token found")

	sess, err := a.sess.Get(c)
	if err != nil {
		return h.Res(c, http.StatusInternalServerError, "Failed to get session")
	}
	a.log.Debug().Msgf("Session ID: %s", sess.ID())
	if c.Cookies("session_id") == "" {
		a.log.Warn().Msg("No session cookie found")
		return h.Res(c, http.StatusUnauthorized, "Not logged in")
	}
	a.log.Trace().Msgf("Session cookie: %s", c.Cookies("session_id"))

	sessionID := sess.ID()
	sessionFallback := sess.Get(c.Cookies("session_id"))
	if sessionID == "" && sessionFallback == nil {
		a.log.Warn().Msg("No session ID found")
		a.log.Trace().Msgf("session keys: %+v", sess.Keys())
		return h.Res(c, http.StatusUnauthorized, "Not logged in")
	}

	if sessionID != c.Cookies("session_id") {
		a.log.Warn().Msg("Session ID mismatch")
		a.log.Debug().Msgf("Session ID: %s", sessionID)
		a.log.Debug().Msgf("Cookie session ID: %s", c.Cookies("session_id"))
		return h.Res(c, http.StatusUnauthorized, "Not logged in")
	}
	a.log.Trace().Msg("Session ID matches cookie")

	memName := sess.Get("member_name")
	if memName == nil {
		a.log.Warn().Msg("No member name found")
		return h.Res(c, http.StatusUnauthorized, "Not logged in")
	}

	a.log.Trace().Msg("should be logged in")
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":         "Logged in",
		"isAuthenticated": true,
		"memberName":      memName,
	})
}

func (a *Service) identifyDevice() (uuid.UUID, error) {
	deviceID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	return deviceID, nil
}
