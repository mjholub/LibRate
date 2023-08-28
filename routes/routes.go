package routes

import (
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers"
	"codeberg.org/mjh/LibRate/controllers/auth"
	"codeberg.org/mjh/LibRate/controllers/form"
	"codeberg.org/mjh/LibRate/controllers/version"
	"codeberg.org/mjh/LibRate/middleware"
	"codeberg.org/mjh/LibRate/models"
	// "codeberg.org/mjh/LibRate/middleware"
)

// Setup handles all the routes for the application
// It receives the configuration, logger and db connection from main
// and then passes them to the controllers
func Setup(logger *zerolog.Logger,
	conf *cfg.Config,
	dbConn *sqlx.DB,
	app *fiber.App,
	fzlog *fiber.Handler,
) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	// setup the middleware
	// NOTE: unsure if this handler is correct
	api := app.Group("/api", *fzlog)
	// initialize the reading of the static files in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		staticPath, err := filepath.Abs("./fe/build")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get absolute path for static files")
			errChan <- err
		}
		assetPath, err := filepath.Abs("./static")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get absolute path for static files")
			errChan <- err
		}

		app.Use("/", filesystem.New(filesystem.Config{
			Root:   http.Dir(staticPath),
			Browse: true,
		}))
		app.Use("/static", filesystem.New(filesystem.Config{
			Root:   http.Dir(assetPath),
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
		mStor = models.NewMemberStorage(dbConn, logger, conf)
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
	formCon := form.NewFormController(logger, *mediaStor)
	sc := controllers.NewSearchController(dbConn)

	app.Get("/api/version", version.Get)

	reviews := api.Group("/reviews")
	reviews.Get("/latest", reviewSvc.GetLatestRatings)
	// TODO: handler for single review based on id
	reviews.Get("/", reviewSvc.GetRatings)
	reviews.Post("/", middleware.Protected(nil), reviewSvc.PostRating)
	reviews.Patch("/:id", middleware.Protected(nil), reviewSvc.UpdateRating)
	reviews.Delete("/:id", middleware.Protected(nil), reviewSvc.DeleteRating)
	// ...or define the GetRatings handler in a way where it returns all ratings if no id is given
	reviews.Get("/:id", reviewSvc.GetRatings)

	authApi := api.Group("/authenticate")
	authApi.Get("/", middleware.Protected(nil), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	member := api.Group("/members")
	member.Post("/login", middleware.Protected(logger), authSvc.Login)
	member.Post("/register", middleware.Protected(logger), authSvc.Register)
	member.Get("/:id", memberSvc.GetMember)

	// NOTE: is protected middleware needed here?
	app.Post("/api/password-entropy", middleware.Protected(nil), auth.ValidatePassword())

	media := api.Group("/media")
	media.Get("/random", mediaCon.GetRandom)
	media.Get("/:id/images", mediaCon.GetImagePaths)
	media.Get("/:id", mediaCon.GetMedia)

	formApi := api.Group("/form")
	// TODO: make the timeouts configurable
	formApi.Post("/add_media/:type", middleware.Protected(nil), timeout.NewWithContext(formCon.AddMedia, 10*time.Second))
	formApi.Post("/update_media/:type", middleware.Protected(nil), formCon.UpdateMedia)

	search := api.Group("/search")
	search.Post("/", sc.Search)
	search.Options("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return nil
}
