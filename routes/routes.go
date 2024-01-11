package routes

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers"
	"codeberg.org/mjh/LibRate/controllers/auth"
	"codeberg.org/mjh/LibRate/controllers/form"
	"codeberg.org/mjh/LibRate/controllers/media"
	memberCtrl "codeberg.org/mjh/LibRate/controllers/members"
	"codeberg.org/mjh/LibRate/controllers/static"
	"codeberg.org/mjh/LibRate/controllers/version"
	"codeberg.org/mjh/LibRate/middleware"
	"codeberg.org/mjh/LibRate/models"
	"codeberg.org/mjh/LibRate/models/member"
)

// Setup handles all the routes for the application
// It receives the configuration, logger and db connection from main
// and then passes them to the controllers
func Setup(
	logger *zerolog.Logger,
	fzlog fiber.Handler,
	conf *cfg.Config,
	dbConn *sqlx.DB,
	newDBConn *pgxpool.Pool,
	app *fiber.App,
	sess *session.Store,
) error {
	api := app.Group("/api", fzlog)

	var (
		mStor     member.MemberStorer
		mediaStor *models.MediaStorage
	)

	switch conf.Engine {
	case "postgres", "sqlite", "mariadb":
		mStor = member.NewSQLStorage(dbConn, logger, conf)
	default:
		return fmt.Errorf("unsupported database engine \"%q\" or error reading config", conf.Engine)
	}
	mediaStor = models.NewMediaStorage(newDBConn, dbConn, logger)

	memberSvc := memberCtrl.NewController(mStor, dbConn, logger, conf)
	formCon := form.NewController(logger, *mediaStor, conf)
	uploadSvc := static.NewController(conf, dbConn, logger)
	sc := controllers.NewSearchController(dbConn)

	app.Get("/api/version", version.Get)

	setupReviews(api, sess, logger, conf, dbConn)

	setupAuth(api, sess, logger, conf, mStor)

	members := api.Group("/members")
	members.Post("/check", memberSvc.Check)
	members.Patch("/update/:member_name", middleware.Protected(sess, logger, conf), memberSvc.Update)
	members.Get("/:email_or_username/info", memberSvc.GetMemberByNickOrEmail)

	setupMedia(api, mediaStor)

	formAPI := api.Group("/form")
	// TODO: make the timeouts configurable
	formAPI.Post("/add_media/:type", middleware.Protected(sess, logger, conf), timeout.NewWithContext(formCon.AddMedia, 10*time.Second))
	formAPI.Post("/update_media/:type", middleware.Protected(sess, logger, conf), formCon.UpdateMedia)

	uploadAPI := api.Group("/upload")
	uploadAPI.Get("/max-file-size", func(c *fiber.Ctx) error { return c.SendString(fmt.Sprintf("%d", conf.Fiber.MaxUploadSize)) })
	uploadAPI.Post("/image", middleware.Protected(sess, logger, conf), uploadSvc.UploadImage)
	uploadAPI.Delete("/image/:id", middleware.Protected(sess, logger, conf), uploadSvc.DeleteImage)

	search := api.Group("/search")
	search.Post("/", sc.Search)
	search.Options("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		if dbConn.Ping() == nil {
			return c.SendStatus(fiber.StatusOK)
		}
		return c.SendStatus(fiber.StatusServiceUnavailable)
	})
	err := setupStatic(app)
	if err != nil {
		return fmt.Errorf("failed to setup static files: %w", err)
	}
	logger.Debug().Msg("static files initialized")

	return nil
}

func setupReviews(api fiber.Router, sess *session.Store, logger *zerolog.Logger, conf *cfg.Config, dbConn *sqlx.DB) {
	rStor := models.NewRatingStorage(dbConn, logger)
	reviewSvc := controllers.NewReviewController(*rStor)

	reviews := api.Group("/reviews")
	reviews.Get("/latest", reviewSvc.GetLatest)
	reviews.Post("/", middleware.Protected(sess, logger, conf), reviewSvc.PostRating)
	reviews.Patch("/:id", middleware.Protected(sess, logger, conf), reviewSvc.UpdateRating)
	reviews.Delete("/:id", middleware.Protected(sess, logger, conf), reviewSvc.DeleteRating)
	reviews.Get("/:media_id", reviewSvc.GetMediaReviews)
	reviews.Get("/:media_id/average", reviewSvc.GetAverageRating)
	reviews.Get("/:id", reviewSvc.GetByID)
}

func setupAuth(
	api fiber.Router,
	sess *session.Store,
	logger *zerolog.Logger,
	conf *cfg.Config,
	mStor member.MemberStorer,
) {
	authSvc := auth.NewService(conf, mStor, logger, sess)

	authAPI := api.Group("/authenticate")
	authAPI.Get("/status", authSvc.GetAuthStatus)
	authAPI.Post("/login", timeout.NewWithContext(authSvc.Login, 10*time.Second))
	authAPI.Post("/logout", authSvc.Logout)
	authAPI.Post("/register", authSvc.Register)
}

func setupMedia(
	api fiber.Router,
	mediaStor *models.MediaStorage,
) {
	mediaCon := media.NewController(*mediaStor)

	media := api.Group("/media")
	media.Get("/random", mediaCon.GetRandom)
	media.Get("/:media_id/images", mediaCon.GetImagePaths)
	media.Get("/:id", mediaCon.GetMedia)
	media.Get("/:media_id/cast", timeout.NewWithContext(mediaCon.GetCastByMediaID, 10*time.Second))
	media.Get("/creator", timeout.NewWithContext(mediaCon.GetCreatorByID, 10*time.Second))
	media.Get("/genres/:kind", timeout.NewWithContext(mediaCon.GetGenres, 30*time.Second))
	media.Get("/genres/:kind/:genre", timeout.NewWithContext(mediaCon.GetGenre, 30*time.Second))
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

		var mu sync.Mutex

		mu.Lock()
		app.Use("/static", filesystem.New(filesystem.Config{
			Root:   http.Dir(assetPath),
			Browse: true,
		}))
		mu.Unlock()
		app.Use("/", filesystem.New(filesystem.Config{
			Root:         http.Dir(staticPath),
			Browse:       false,
			NotFoundFile: "404.html",
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
