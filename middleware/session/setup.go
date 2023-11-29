package session

import (
	"time"

	sess "github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofrs/uuid/v5"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/internal/crypt"
)

func Setup(conf *cfg.Config, cryptStore *crypt.Storage) (*sess.Store, error) {
	return sess.New(
		sess.Config{
			Storage:           cryptStore,
			Expiration:        7 * 24 * time.Hour,
			KeyLookup:         "cookie:session_id",
			KeyGenerator:      uuid.Must(uuid.NewV7()).String,
			CookieDomain:      conf.Fiber.Domain,
			CookieHTTPOnly:    true,
			CookieSessionOnly: true,
			CookieSecure:      true,
		},
	), nil
}
