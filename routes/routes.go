package routes

import (
	"net/http"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers"
	"codeberg.org/mjh/LibRate/controllers/auth"
	"codeberg.org/mjh/LibRate/controllers/version"
	"codeberg.org/mjh/LibRate/models"
	// "codeberg.org/mjh/LibRate/middleware"
)

// Setup handles all the routes for the application
// It receives the configuration, logger and db connection from main
// and then passes them to the controllers
func Setup(logger *zerolog.Logger, conf *cfg.Config, dbConn *sqlx.DB, app *fiber.App) {
	staticPath, err := filepath.Abs("./fe/build")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get absolute path for static files")
	}

	app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.Dir(staticPath),
		Browse: true,
	}))

	mStor := models.NewMemberStorage(dbConn, logger)
	authSvc := auth.NewAuthService(conf, mStor)
	rStor := models.NewRatingStorage(dbConn, logger)
	reviewSvc := controllers.NewReviewController(*rStor)
	memberSvc := controllers.NewMemberController(*mStor)

	app.Get("/api/version", version.Get)
	app.Get("/api/reviews/:id", reviewSvc.GetRatings)
	app.Get("/api/reviews/", reviewSvc.GetRatings)
	app.Get("/api/reviews/latest", reviewSvc.GetLatestRatings)
	app.Get("/api/member/:id", memberSvc.GetMember)
	app.Get("api/reviews/:mediaID", reviewSvc.GetRatings)
	app.Post("/api/password-entropy", auth.ValidatePassword())
	app.Post("/api/reviews", reviewSvc.PostRating)
	app.Post("/api/login", authSvc.Login)
	app.Post("/api/register", authSvc.Register)
	app.Post("/api/search", controllers.SearchMedia)
	app.Options("/api/search", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
}
