package routes

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/jmoiron/sqlx"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers"
	"codeberg.org/mjh/LibRate/controllers/auth"
	"codeberg.org/mjh/LibRate/controllers/form"
	"codeberg.org/mjh/LibRate/controllers/media"
	memberCtrl "codeberg.org/mjh/LibRate/controllers/members"
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
	conf *cfg.Config,
	dbConn *sqlx.DB,
	neo4jConn *neo4j.DriverWithContext,
	app *fiber.App,
	fzlog *fiber.Handler,
	sqlcipher *sql.DB,
) error {
	// setup the middleware
	// NOTE: unsure if this handler is correct
	api := app.Group("/api", *fzlog)

	var (
		mStor     member.MemberStorer
		rStor     *models.RatingStorage
		mediaStor *models.MediaStorage
	)

	switch conf.Engine {
	case "postgres", "sqlite", "mariadb":
		mStor = member.NewSQLStorage(dbConn, logger, conf)
	case "neo4j":
		mStor = member.NewNeo4jStorage(*neo4jConn, logger, conf)
	default:
		return fmt.Errorf("unsupported database engine \"%q\" or error reading config", conf.Engine)
	}
	rStor = models.NewRatingStorage(dbConn, logger)
	mediaStor = models.NewMediaStorage(dbConn, logger)

	authSvc := auth.NewService(conf, mStor, logger, sqlcipher)
	reviewSvc := controllers.NewReviewController(*rStor)
	memberSvc := memberCtrl.NewController(mStor, logger, conf)
	mediaCon := media.NewController(*mediaStor)
	formCon := form.NewController(logger, *mediaStor, conf)
	sc := controllers.NewSearchController(dbConn)

	app.Get("/api/version", version.Get)
	// TODO: add template rendering
	// app.Get("/noscript",

	reviews := api.Group("/reviews")
	reviews.Get("/latest", reviewSvc.GetLatest)
	reviews.Post("/", middleware.Protected(nil, conf), reviewSvc.PostRating)
	reviews.Patch("/:id", middleware.Protected(nil, conf), reviewSvc.UpdateRating)
	reviews.Delete("/:id", middleware.Protected(nil, conf), reviewSvc.DeleteRating)
	reviews.Get("/:media_id", reviewSvc.GetMediaReviews)
	reviews.Get("/:media_id/average", reviewSvc.GetAverageRating)
	reviews.Get("/:id", reviewSvc.GetByID)

	authAPI := api.Group("/authenticate")
	authAPI.Get("/", middleware.Protected(nil, conf), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	members := api.Group("/members")
	members.Post("/login", timeout.NewWithContext(authSvc.Login, 10*time.Second))
	members.Post("/register", authSvc.Register)
	members.Post("/check", memberSvc.Check)
	members.Get("/:nickname/info", memberSvc.GetMemberByNick)
	members.Get("/id/:nickname", memberSvc.GetID)
	// pubkey returns a single use public key for the client to encrypt their password with
	// this is to prevent the server from ever knowing the user's password
	members.Get("pubkey", authSvc.GetPubKey)

	media := api.Group("/media")
	media.Get("/random", mediaCon.GetRandom)
	media.Get("/:media_id/images", mediaCon.GetImagePaths)
	media.Get("/:id", mediaCon.GetMedia)
	media.Get("/:media_id/cast", timeout.NewWithContext(mediaCon.GetCastByMediaID, 10*time.Second))
	media.Get("/creator", timeout.NewWithContext(mediaCon.GetCreatorByID, 10*time.Second))

	formAPI := api.Group("/form")
	// TODO: make the timeouts configurable
	formAPI.Post("/add_media/:type", middleware.Protected(nil, conf), timeout.NewWithContext(formCon.AddMedia, 10*time.Second))
	formAPI.Post("/update_media/:type", middleware.Protected(nil, conf), formCon.UpdateMedia)

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

	app.Get("/api/health", func(c *fiber.Ctx) error {
		if dbConn.Ping() == nil {
			return c.SendStatus(fiber.StatusOK)
		}
		return c.SendStatus(fiber.StatusServiceUnavailable)
	})

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
