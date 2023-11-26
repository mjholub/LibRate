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

	"codeberg.org/mjh/LibRate/cfg"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/earlydata"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/storage/redis/v3"
	"github.com/gofiber/template/html/v2"
)

func CreateApp(renderEngine *html.Engine, conf *cfg.Config) *fiber.App {
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

func SetupMiddlewares(conf *cfg.Config, logger *zerolog.Logger) []fiber.Handler {
	return []fiber.Handler{
		idempotency.New(),
		helmet.New(),
		csrf.New(csrf.Config{
			// FIXME: stupid svelte won't load X-CSRF-Token from cookies
			KeyLookup:         "cookie:csrf_",
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
				Database: conf.Redis.Database + 1,
			}),
		}),
		recover.New(),
		earlydata.New(),
		cache.New(cache.Config{
			Expiration: 15 * time.Minute,
			Storage: redis.New(redis.Config{
				Host:     conf.Redis.Host,
				Port:     conf.Redis.Port,
				Username: conf.Redis.Username,
				Password: conf.Redis.Password,
				Database: conf.Redis.Database + 2,
			}),
			Next: func(c *fiber.Ctx) bool {
				return c.Query("cache") == "false"
			},
		}),
		compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}),
		setupLogger(logger),
	}
}

func setupLogger(logger *zerolog.Logger) fiber.Handler {
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
