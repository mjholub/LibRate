package cmd

import (
	//"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/middleware/render"
	"codeberg.org/mjh/LibRate/middleware/security"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	//"github.com/gofiber/fiber/v2/middleware/earlydata"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/storage/redis/v3"
)

func CreateApp(conf *cfg.Config) *fiber.App {
	renderEngine := render.Setup(conf)

	tag, err := getLatestTag()
	if err != nil {
		tag = "unknown"
	}
	_ = tag

	app := fiber.New(fiber.Config{
		AppName: "LibRate v0.8.12",
		// FIXME: earlydata not working with containers
		// EnableTrustedProxyCheck: true,
		// TrustedProxies:          []string{"127.0.0.1", "::1", "localhost", conf.Fiber.Host, conf.Fiber.Domain},
		Prefork:           conf.Fiber.Prefork,
		ReduceMemoryUsage: conf.Fiber.ReduceMemUsage,
		Views:             renderEngine,
		JSONEncoder:       json.Marshal,
		JSONDecoder:       json.Unmarshal,
	},
	)

	return app
}

func SetupMiddlewares(conf *cfg.Config,
	logger *zerolog.Logger,
) []fiber.Handler {
	return []fiber.Handler{
		idempotency.New(idempotency.Config{
			Next: func(c *fiber.Ctx) bool {
				return lo.Contains([]string{"/api/authenticate", "/api/media/random", "/api/media/genres"}, c.Path())
			},
		}),
		security.SetupHelmet(conf),
		security.SetupCSRF(conf, logger),
		security.SetupCORS(conf),
		recover.New(),
		// earlydata.New(),
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
				return c.Query("cache") == "false" || c.Path() == "/api/authenticate/status"
			},
		}),
		compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}),
	}
}

func SetupLogger(conf *cfg.Config, logger *zerolog.Logger) fiber.Handler {
	fiberlog := fiberzerolog.New(fiberzerolog.Config{
		Logger: logger,
		// skip logging for static files, there's too many of them
		Next: func(c *fiber.Ctx) bool {
			if conf.Logging.Level != "trace" {
				return strings.Contains(c.Path(), "/_app/")
			}
			return false
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
