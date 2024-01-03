package session

import (
	"time"

	sess "github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/postgres/v2"
	"github.com/gofrs/uuid/v5"

	"codeberg.org/mjh/LibRate/cfg"
)

func Setup(conf *cfg.Config) (*sess.Store, error) {
	sessionStorage := postgres.New(postgres.Config{
		Host:       conf.DBConfig.Host,
		Port:       int(conf.DBConfig.Port),
		Database:   "librate_sessions",
		Table:      "sessions",
		Username:   conf.DBConfig.User,
		Password:   conf.Secret,
		GCInterval: 30 * time.Minute,
	})
	return sess.New(
		sess.Config{
			Storage:           sessionStorage,
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
