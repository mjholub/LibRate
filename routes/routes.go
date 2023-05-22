package routes

import (
	"net/http"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	"codeberg.org/mjh/LibRate/controllers"
	"codeberg.org/mjh/LibRate/controllers/auth"
	"codeberg.org/mjh/LibRate/middleware"
	"codeberg.org/mjh/LibRate/utils"
)

func Setup(app *fiber.App) {
	log := utils.NewLogger()
	staticPath, err := filepath.Abs("./fe/public")
	if err != nil {
		log.Sugar().Fatalf("Error loading static path: %s", err)
	}

	app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.Dir(staticPath),
		Browse: true,
	}))

	app.Get("/api/reviews/:id", controllers.GetRatings)
	app.Patch("/api/password-entropy", middleware.Protected(), auth.ValidatePassword())
	app.Post("/api/reviews", controllers.PostRating)
	app.Post("/api/login", auth.Login)
	app.Post("/api/register", auth.Register)
}
