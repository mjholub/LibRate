package security

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/storage/redis/v3"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
)

func SetupCSRF(conf *cfg.Config, logger *zerolog.Logger) fiber.Handler {
	return csrf.New(csrf.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			path := c.Path()
			if conf.LibrateEnv == "development" {
				switch err {
				case csrf.ErrTokenNotFound:
					logger.Warn().Str("path", path).Msg("CSRF token not found")
				case csrf.ErrBadReferer:
					logger.Warn().Str("path", path).Msg("CSRF bad referer")
				case csrf.ErrNoReferer:
					logger.Warn().Str("path", path).Msg("CSRF no referer")
				case csrf.ErrTokenInvalid:
					logger.Warn().Str("path", path).Msgf("invalid CSRF token: %s", c.Cookies("csrf_"))
				default:
					logger.Warn().Str("path", path).Msgf("CSRF error: %s", err.Error())
				}
				return c.Status(fiber.StatusForbidden).Render("error", fiber.Map{
					"Title":   "Forbidden",
					"Status":  fiber.StatusForbidden,
					"Message": "I see what you did there.",
				}, "error")
			}
			return c.Next()
		},
		KeyLookup:         "header:X-CSRF-Token",
		CookieName:        "csrf_",
		CookieSessionOnly: true,
		CookieSecure:      true,
		CookieSameSite:    "Lax",
		Expiration:        2 * time.Hour,
		KeyGenerator:      uuid.Must(uuid.NewV4()).String,
		Storage: redis.New(redis.Config{
			Host:     conf.Redis.Host,
			Port:     conf.Redis.Port,
			Username: conf.Redis.Username,
			Password: conf.Redis.Password,
			Database: conf.Redis.CsrfDB,
		}),
	})
}
