package routes

import (
	"net/http"
	"path/filepath"
	"sync"

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
func Setup(logger *zerolog.Logger, conf *cfg.Config, dbConn *sqlx.DB, app *fiber.App) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	// initialize the reading of the static files in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		staticPath, err := filepath.Abs("./fe/build")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get absolute path for static files")
			errChan <- err
		}

		app.Use("/", filesystem.New(filesystem.Config{
			Root:   http.Dir(staticPath),
			Browse: true,
		}))
	}()

	var (
		mStor     *models.MemberStorage
		rStor     *models.RatingStorage
		mediaStor *models.MediaStorage
	)
	wg.Add(3)

	go func() {
		defer wg.Done()
		mStor = models.NewMemberStorage(dbConn, logger)
	}()
	go func() {
		defer wg.Done()
		rStor = models.NewRatingStorage(dbConn, logger)
	}()
	go func() {
		defer wg.Done()
		mediaStor = models.NewMediaStorage(dbConn, logger)
	}()

	// wait for the static files to be read and the data layer to be initialized
	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		for err := range errChan {
			if err != nil {
				logger.Error().Err(err).Msg("Error reading static files")
				return err
			}
		}
	}

	authSvc := auth.NewAuthService(conf, mStor)
	reviewSvc := controllers.NewReviewController(*rStor)
	memberSvc := controllers.NewMemberController(*mStor)
	mediaCon := controllers.NewMediaController(*mediaStor)
	sc := controllers.NewSearchController(dbConn)

	app.Get("/api/version", version.Get)

	app.Get("/api/reviews/:id", reviewSvc.GetRatings)
	app.Get("/api/reviews/", reviewSvc.GetRatings)
	app.Post("/api/reviews", reviewSvc.PostRating)
	app.Get("/api/reviews/latest", reviewSvc.GetLatestRatings)
	app.Get("api/reviews/:mediaID", reviewSvc.GetRatings)

	app.Post("/api/login", authSvc.Login)
	app.Post("/api/register", authSvc.Register)
	app.Get("/api/member/:id", memberSvc.GetMember)
	app.Post("/api/password-entropy", auth.ValidatePassword())

	app.Get("/api/media/random", mediaCon.GetRandom)
	app.Get("/api/media/:id", mediaCon.GetMedia)

	app.Post("/api/search", sc.Search)
	app.Options("/api/search", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return nil
}
