package routes

import (
	"net/http"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	"codeberg.org/mjh/LibRate/controllers"
	"codeberg.org/mjh/LibRate/controllers/auth"
	"codeberg.org/mjh/LibRate/internal/logging"
	// "codeberg.org/mjh/LibRate/middleware"
)

func Setup(app *fiber.App) {
	lrlog := logging.Init()
	staticPath, err := filepath.Abs("./fe/public")
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
}
