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

func (a *Service) createSession(c *fiber.Ctx, rememberMe bool, member *member.Member) error {
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
	if sess == nil {
		a.log.Error().Msg("Failed to create session: session is nil")
		return h.Res(c, http.StatusInternalServerError, "Failed to create session")
	}
	a.log.Debug().Msgf("Creating session with ID: %s", sess.ID())
	var mu sync.Mutex

	mu.Lock()
	go sess.Set("member_name", member.MemberName)
	go sess.Set("session_id", sess.ID())
	go sess.Set("device_id", deviceHash)
	go sess.Set("ip", c.IP())
	go sess.Set("user_agent", string(c.Request().Header.UserAgent()))
	timeout, err := a.GetSessionTimeoutPrefs(rememberMe)
	if err != nil {
		return err
	}
	sess.SetExpiry(timeout)

	signedToken, err := a.createToken(member, &timeout, sess)
	if err != nil {
		a.log.Err(err)
		return h.Res(c, fiber.StatusInternalServerError, "Failed to prepare session")
	}
	c.Cookie(&fiber.Cookie{
		HTTPOnly: true,
		Name:     "session_id",
		MaxAge:   int(timeout.Seconds()),
		Domain:   a.conf.Fiber.Domain,
		SameSite: "Lax",
		Value:    sess.ID(),
	},
	)

	mu.Unlock()

	if err = sess.Save(); err != nil {
		a.log.Error().Err(err).Msgf("Failed to create session: %s", err.Error())
		return h.Res(c, http.StatusInternalServerError, "Failed to create session")
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":    "Logged in successfully",
		"token":      signedToken,
		"memberName": member.MemberName,
	})
}

func (a *Service) createToken(member *member.Member, timeout *time.Duration, sess *session.Session) (t string, err error) {
	// TODO: add role checking that works with pq.StringArray
	claims := jwt.MapClaims{
		"member_name": member.MemberName,
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
	a.log.Debug().Msg("CSRF token found")

	sess, err := a.sess.Get(c)
	if err != nil {
		return h.Res(c, http.StatusInternalServerError, "Failed to get session")
	}
	a.log.Debug().Msgf("Session ID: %s", sess.ID())
	if c.Cookies("session_id") == "" {
		a.log.Warn().Msg("No session cookie found")
		return h.Res(c, http.StatusUnauthorized, "Not logged in")
	}
	a.log.Debug().Msgf("Session cookie: %s", c.Cookies("session_id"))

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
	a.log.Debug().Msg("Session ID matches cookie")

	memName := sess.Get("member_name")
	if memName == nil {
		a.log.Warn().Msg("No member name found")
		return h.Res(c, http.StatusUnauthorized, "Not logged in")
	}

	a.log.Debug().Msg("should be logged in")
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":         "Logged in",
		"isAuthenticated": true,
		"memberName":      memName,
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
// Currently, only support for "remember me" is implemented
func (a *Service) GetSessionTimeoutPrefs(rememberMe bool) (timeout time.Duration, err error) {
	if rememberMe {
		return time.ParseDuration("2160h") // 90 days
	}
	return time.ParseDuration("1h")
}
