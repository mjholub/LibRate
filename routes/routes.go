package routes

import (
	"net/http"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers"
	"codeberg.org/mjh/LibRate/controllers/auth"
	"codeberg.org/mjh/LibRate/internal/logging"
	// "codeberg.org/mjh/LibRate/middleware"
)

func Setup(app *fiber.App) {
	logConf := cfg.LoadLoggerConfig().OrElse(func(err error) logging.Config {
		return logging.Config{
			Level:  "info",
			Target: "stdout",
			Format: "json",
		}
	}(nil))
	lrlog := logging.Init(&logConf)
	staticPath, err := filepath.Abs("./fe/build")
	if err != nil {
		lrlog.Error().Err(err).Msg("Failed to get absolute path for static files")
	}

	app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.Dir(staticPath),
		Browse: true,
	}))

	app.Get("/api/reviews/:id", controllers.GetRatings)
	app.Post("/api/password-entropy", auth.ValidatePassword())
	app.Post("/api/reviews", controllers.PostRating)
	app.Post("/api/login", auth.Login)
	app.Post("/api/register", auth.Register)
	app.Post("/api/search", controllers.SearchMedia)
	app.Options("/api/search", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
}
