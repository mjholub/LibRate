package routes

import (
	"fmt"
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
	"codeberg.org/mjh/LibRate/controllers/media"
	"codeberg.org/mjh/LibRate/controllers/members"
	"codeberg.org/mjh/LibRate/controllers/version"
	"codeberg.org/mjh/LibRate/middleware"
	"codeberg.org/mjh/LibRate/models"
)

// Setup handles all the routes for the application
// It receives the configuration, logger and db connection from main
// and then passes them to the controllers
func Setup(
	logger *zerolog.Logger,
	conf *cfg.Config,
	dbConn *sqlx.DB,
	app *fiber.App,
	fzlog *fiber.Handler,
) error {
	// setup the middleware
	// NOTE: unsure if this handler is correct
	api := app.Group("/api", *fzlog)

	var (
		mStor     *models.MemberStorage
		rStor     *models.RatingStorage
		mediaStor *models.MediaStorage
	)

	mStor = models.NewMemberStorage(dbConn, logger, conf)
	rStor = models.NewRatingStorage(dbConn, logger)
	mediaStor = models.NewMediaStorage(dbConn, logger)

	authSvc := auth.NewAuthService(conf, mStor, logger)
	reviewSvc := controllers.NewReviewController(*rStor)
	memberSvc := members.NewController(mStor, logger)
	mediaCon := media.NewController(*mediaStor)
	formCon := form.NewFormController(logger, *mediaStor)
	sc := controllers.NewSearchController(dbConn)

	app.Get("/api/version", version.Get)

	reviews := api.Group("/reviews")
	reviews.Get("/latest", reviewSvc.GetLatestRatings)
	// TODO: handler for single review based on id
	reviews.Post("/", middleware.Protected(nil, conf), reviewSvc.PostRating)
	reviews.Patch("/:id", middleware.Protected(nil, conf), reviewSvc.UpdateRating)
	reviews.Delete("/:id", middleware.Protected(nil, conf), reviewSvc.DeleteRating)
	// ...or define the GetRatings handler in a way where it returns all ratings if no id is given
	reviews.Get("/:id", reviewSvc.GetRatings)

	authApi := api.Group("/authenticate")
	authApi.Get("/", middleware.Protected(nil, conf), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	member := api.Group("/members")
	member.Post("/login", timeout.NewWithContext(authSvc.Login, 10*time.Second))
	member.Post("/register", authSvc.Register)
	member.Get("/:id", memberSvc.GetMember)
	member.Get("/:nickname/info", memberSvc.GetMemberByNick)

	app.Post("/api/password-entropy", auth.ValidatePassword())

	media := api.Group("/media")
	media.Get("/random", mediaCon.GetRandom)
	media.Get("/:id/images", mediaCon.GetImagePaths)
	media.Get("/:id", mediaCon.GetMedia)

	formApi := api.Group("/form")
	// TODO: make the timeouts configurable
	formApi.Post("/add_media/:type", middleware.Protected(nil, conf), timeout.NewWithContext(formCon.AddMedia, 10*time.Second))
	formApi.Post("/update_media/:type", middleware.Protected(nil, conf), formCon.UpdateMedia)

	search := api.Group("/search")
	search.Post("/", sc.Search)
	search.Options("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	err := setupStatic(app)
	if err != nil {
		return fmt.Errorf("failed to setup static files: %w", err)
	}
	logger.Debug().Msg("static files initialized")
	return nil
}

func setupStatic(app *fiber.App) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		staticPath, err := filepath.Abs("./fe/build/")
		if err != nil {
			errChan <- fmt.Errorf("failed to get absolute path for static files: %w", err)
		}
		assetPath, err := filepath.Abs("./static")
		if err != nil {
			errChan <- fmt.Errorf("failed to get absolute path for static files: %w", err)
		}

		app.Use("/", filesystem.New(filesystem.Config{
			Root:         http.Dir(staticPath),
			Browse:       true,
			NotFoundFile: "404.html",
		}))
		app.Use("/static", filesystem.New(filesystem.Config{
			Root:   http.Dir(assetPath),
			Browse: true,
		}))
	}()

	// wait for the static files to be read and the data layer to be initialized
	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		for err := range errChan {
			if err != nil {
				return fmt.Errorf("error reading static files: %w", err)
			}
		}
	}

	return nil
}
