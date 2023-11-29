package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"codeberg.org/mjh/LibRate/cfg"
	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/middleware/render"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/earlydata"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
)

func CreateApp(conf *cfg.Config) *fiber.App {
	renderEngine := render.Setup(conf)

	tag, err := getLatestTag()
	if err != nil {
		tag = "unknown"
	}

	app := fiber.New(fiber.Config{
		AppName:                 fmt.Sprintf("LibRate %s", tag),
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"127.0.0.1", "::1", "localhost", conf.Fiber.Host},
		Prefork:                 conf.Fiber.Prefork,
		ReduceMemoryUsage:       conf.Fiber.ReduceMemUsage,
		Views:                   renderEngine,
		JSONEncoder:             json.Marshal,
		JSONDecoder:             json.Unmarshal,
	},
	)

	return app
}

func SetupMiddlewares(conf *cfg.Config,
	logger *zerolog.Logger, session *session.Store,
) []fiber.Handler {
	fh := conf.Fiber.Host
	fp := conf.Fiber.Port
	localAliases := strings.ReplaceAll(fmt.Sprintf(`%s:%d https://%s:%d http://%s:%d https://lr.localhost`,
		fh, fp, fh, fp, fh, fp), "'", "")
	return []fiber.Handler{
		idempotency.New(idempotency.Config{
			Next: func(c *fiber.Ctx) bool {
				return lo.Contains([]string{"/api/authenticate", "/api/media/random"}, c.Path())
			},
		}),
		helmet.New(helmet.Config{
			XSSProtection:  "1; mode=block",
			ReferrerPolicy: "no-referrer-when-downgrade",
			ContentSecurityPolicy: fmt.Sprintf(`default-src 'self' https://gnu.org https://www.gravatar.com %s;
				style-src 'self' cdn.jsdelivr.net 'unsafe-inline';
				script-src 'self' https://unpkg.com/htmx.org@1.9.9 %s 'unsafe-inline' 'unsafe-eval';
				img-src 'self' https://www.gravatar.com data:;`,
				localAliases, localAliases),
		}),
		csrf.New(csrf.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				accepts := c.Accepts("form/multipart", "application/json")
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
					if strings.Contains(accepts, string(c.Request().Header.Peek("Accept"))) {
						return h.Res(c, fiber.StatusForbidden, "Forbidden")
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
		}),
		cors.New(cors.Config{
			AllowOrigins: fmt.Sprintf(
				`%s, %s, https://gnu.org, https://gravatar.com, https://unpkg.com/htmx.org@latest`,
				conf.Fiber.Domain, conf.Fiber.Host),
			AllowHeaders: "Origin, Content-Type, Accept, X-Requested-With, X-Csrf-Token",
			AllowMethods: "GET, POST, PUT, DELETE, HEAD, PATCH, OPTIONS",
		}),
		recover.New(),
		earlydata.New(),
		etag.New(),
		cache.New(cache.Config{
			Expiration: 15 * time.Minute,
			Storage: redis.New(redis.Config{
				Host:     conf.Redis.Host,
				Port:     conf.Redis.Port,
				Username: conf.Redis.Username,
				Password: conf.Redis.Password,
				Database: conf.Redis.CacheDB,
			}),
			Next: func(c *fiber.Ctx) bool {
				return c.Query("cache") == "false"
			},
		}),
		compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}),
		SetupLogger(logger),
	}
}

func SetupLogger(logger *zerolog.Logger) fiber.Handler {
	fiberlog := fiberzerolog.New(fiberzerolog.Config{
		Logger: logger,
		// skip logging for static files, there's too many of them
		Next: func(c *fiber.Ctx) bool {
			return strings.Contains(c.Path(), "/_app/")
		},
	})
	return fiberlog
}

func getLatestTag() (string, error) {
	// check if git is present, otherwise try os.GetEnv("GIT_TAG")
	if _, err := exec.LookPath("git"); err != nil {
		if tag := strings.TrimSpace(os.Getenv("GIT_TAG")); tag != "" {
			return tag, nil
		}
		return "", err
	}
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	latestTag := strings.TrimSpace(string(out))
	return latestTag, nil
}
