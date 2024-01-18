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
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	rec "github.com/gofiber/fiber/v2/middleware/recover"
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
		AppName:           "LibRate v0.8.17", // TODO: add some shell script to generate this on go build
		Prefork:           conf.Fiber.Prefork,
		ReduceMemoryUsage: conf.Fiber.ReduceMemUsage,
		Views:             renderEngine,
		JSONEncoder:       json.Marshal,
		JSONDecoder:       json.Unmarshal,
	},
	)

	return app
}

// SetupWS MUST be called before SetupRoutes
// Otherwise the websocket handler will not be 'shadowed'
// by the static file handler
func SetupWS(app *fiber.App, routes ...string) websocket.Config {
	for i := range routes {
		app.Use("api"+routes[i]+"/ws", func(c *fiber.Ctx) error {
			if websocket.IsWebSocketUpgrade(c) {
				c.Locals("allowed", true)
				return c.Next()
			}
			return fiber.ErrUpgradeRequired
		})
	}

	cfg := websocket.Config{
		RecoverHandler: func(conn *websocket.Conn) {
			if err := recover(); err != nil {
				conn.WriteJSON(fiber.Map{
					"error": err,
				})
			}
		},
	}
	return cfg
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
		// security.SetupCORS(conf),
		rec.New(rec.Config{
			Next: func(c *fiber.Ctx) bool {
				return strings.Contains(c.Route().Path, "/ws")
			},
		}),
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
				return c.Query("cache") == "false" || c.Path() == "/api/authenticate/status" || strings.Contains(c.Route().Path, "/ws")
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
