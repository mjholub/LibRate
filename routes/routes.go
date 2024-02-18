package routes

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers"
	"codeberg.org/mjh/LibRate/controllers/auth"
	"codeberg.org/mjh/LibRate/controllers/form"
	"codeberg.org/mjh/LibRate/controllers/media"
	memberCtrl "codeberg.org/mjh/LibRate/controllers/members"
	"codeberg.org/mjh/LibRate/controllers/search"
	"codeberg.org/mjh/LibRate/controllers/static"
	"codeberg.org/mjh/LibRate/controllers/version"
	"codeberg.org/mjh/LibRate/middleware"
	"codeberg.org/mjh/LibRate/models"
	mediaModels "codeberg.org/mjh/LibRate/models/media"
	"codeberg.org/mjh/LibRate/models/member"
	searchdb "codeberg.org/mjh/LibRate/models/search"
)

type RouterProps struct {
	Conf            *cfg.Config
	Log             *zerolog.Logger
	LogHandler      fiber.Handler
	LegacyDB        *sqlx.DB
	DB              *pgxpool.Pool
	App             *fiber.App
	SessionHandler  *session.Store
	WebsocketConfig *websocket.Config
	Validation      *validator.Validate
	Cache           *redis.Storage
}

// Setup handles all the routes for the application
// It receives the configuration, logger and db connection from main
// and then passes them to the controllers
func Setup(ctx context.Context, r *RouterProps) error {
	api := r.App.Group("/api", r.LogHandler)

	r.App.Get("/docs/*", swagger.New(swagger.Config{
		URL: "/static/meta/swagger.json",
		// TODO: figure out how to use https://github.com/svmk/swagger-i18n-extension#readme
		// with this middleware
		Plugins: []template.JS{
			template.JS("SwaggerUIBundle.plugins.DownloadUrl"),
		},
	}))

	var (
		mStor     member.Storer
		mediaStor *mediaModels.Storage
	)

	switch r.Conf.Engine {
	case "postgres", "sqlite", "mariadb":
		mStor = member.NewSQLStorage(r.LegacyDB, r.DB, r.Log, r.Conf)
	default:
		return fmt.Errorf("unsupported database engine \"%q\" or error reading r.Config", r.Conf.Engine)
	}
	mediaStor = mediaModels.NewStorage(r.DB, r.LegacyDB, r.Log)

	memberSvc := memberCtrl.NewController(mStor, r.LegacyDB, r.SessionHandler, r.Log, r.Conf)
	formCon := form.NewController(r.Log, *mediaStor, r.Conf)
	uploadSvc := static.NewController(r.Conf, r.LegacyDB, r.Log)

	r.App.Get("/api/version", version.Get)

	setupReviews(api, r.SessionHandler, r.Log, r.Conf, r.LegacyDB)

	setupAuth(api, r.SessionHandler, r.Log, r.Conf, mStor)

	setupMembers(memberSvc, api, r.SessionHandler, r.Log, r.Conf)

	setupMedia(api, mediaStor, r.Conf)

	// don't see a point encapsulating 2-3 routes in a separate function
	formAPI := api.Group("/form")
	formAPI.Post("/add_media/:type", middleware.Protected(r.SessionHandler, r.Log, r.Conf), timeout.NewWithContext(formCon.AddMedia, 10*time.Second))
	formAPI.Post("/update_media/:type", middleware.Protected(r.SessionHandler, r.Log, r.Conf), formCon.UpdateMedia)

	setupUpload(uploadSvc, api, r.SessionHandler, r.Log, r.Conf)

	setupSearch(ctx, r.Validation, &r.Conf.CouchDB, r.Cache, r.Log, api)

	r.App.Get("/api/health", func(c *fiber.Ctx) error {
		return c.SendString("I'm alive!")
	})
	err := setupStatic(r.App, r.Conf.Fiber.StaticDir, r.Conf.Fiber.FrontendDir)
	if err != nil {
		return fmt.Errorf("failed to setup static files: %w", err)
	}
	r.Log.Debug().Msg("static files initialized")

	return nil
}

func setupSearch(ctx context.Context, v *validator.Validate, conf *cfg.Search, cache *redis.Storage, log *zerolog.Logger, api fiber.Router) {
	ss, err := searchdb.Connect(conf, log)
	if err != nil {
		log.Err(err).Msgf("an error occured while setting up search handler. Search won't work!")
		return
	}

	searchAPI := api.Group("/search")
	svc, err := search.NewService(
		ctx, v, ss, conf.MainIndexPath, cache, log).Get()
	if err != nil {
		log.Warn().Err(err).Msg("failed to set up routes for search API")
		searchAPI.Post("/", sendNotImpl)
		searchAPI.Get("/", sendNotImpl)
	} else {
		searchAPI.Post("/", svc.HandleSearch)
		searchAPI.Get("/", svc.HandleSearch)
	}
}

func sendNotImpl(c *fiber.Ctx) error {
	return c.Redirect("https://http.cat/images/501.jpg", 303)
}

func setupUpload(uploadSvc *static.Controller, api fiber.Router, sess *session.Store, logger *zerolog.Logger, conf *cfg.Config) {
	uploadAPI := api.Group("/upload")
	uploadAPI.Get("/max-file-size", func(c *fiber.Ctx) error { return c.SendString(fmt.Sprintf("%d", conf.Fiber.MaxUploadSize)) })
	uploadAPI.Post("/image", middleware.Protected(sess, logger, conf), uploadSvc.UploadImage)
	uploadAPI.Delete("/image/:id", middleware.Protected(sess, logger, conf), uploadSvc.DeleteImage)
}

func setupMembers(memberSvc *memberCtrl.Controller, api fiber.Router, sess *session.Store, logger *zerolog.Logger, conf *cfg.Config) {
	members := api.Group("/members")
	members.Post("/check", memberSvc.Check)
	members.Patch("/update/:member_name", middleware.Protected(sess, logger, conf), memberSvc.Update)
	members.Patch("/update/:memeber_name/preferences", middleware.Protected(sess, logger, conf), memberSvc.UpdatePrefs)
	members.Post("/:uuid/ban", middleware.Protected(sess, logger, conf), memberSvc.Ban)
	members.Post("/follow", middleware.Protected(sess, logger, conf), memberSvc.Follow)
	members.Put("/follow/requests/in/:id", middleware.Protected(sess, logger, conf), memberSvc.AcceptFollow)
	members.Delete("/follow/requests/in/:id", middleware.Protected(sess, logger, conf), memberSvc.RejectFollow)
	members.Delete("/follow/requests/out/:id", middleware.Protected(sess, logger, conf), memberSvc.CancelFollowRequest)
	members.Get("/follow/requests/:type", middleware.Protected(sess, logger, conf), memberSvc.GetFollowRequests)
	members.Get("/follow/status/:followee_webfinger", middleware.Protected(sess, logger, conf), memberSvc.FollowStatus)
	members.Delete("/follow", middleware.Protected(sess, logger, conf), memberSvc.Unfollow)
	members.Delete("/:uuid/ban", middleware.Protected(sess, logger, conf), memberSvc.Unban)
	members.Get("/:email_or_username/info", memberSvc.GetMemberByNickOrEmail)
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
	mStor member.Storer,
) {
	authSvc := auth.NewService(conf, mStor, logger, sess)

	authAPI := api.Group("/authenticate")
	authAPI.Post("/login", timeout.NewWithContext(authSvc.Login, 10*time.Second))
	authAPI.Post("/delete-account", middleware.Protected(sess, logger, conf), authSvc.DeleteAccount)
	authAPI.Patch("/password", middleware.Protected(sess, logger, conf), authSvc.ChangePassword)
	authAPI.Get("/status", authSvc.GetAuthStatus)
	authAPI.Post("/logout", authSvc.Logout)
	authAPI.Post("/register", authSvc.Register)
}

func setupMedia(
	api fiber.Router,
	mediaStor *mediaModels.Storage,
	conf *cfg.Config,
) {
	mediaCon := media.NewController(*mediaStor, conf)

	mediaRouter := api.Group("/media")
	mediaRouter.Get("/random", mediaCon.GetRandom)
	mediaRouter.Get("/import-sources", mediaCon.GetImportSources)
	mediaRouter.Get("/:media_id/images", mediaCon.GetImagePaths)
	mediaRouter.Get("/:id", mediaCon.GetMedia)
	mediaRouter.Get("/:media_id/cast", timeout.NewWithContext(mediaCon.GetCastByMediaID, 10*time.Second))
	mediaRouter.Get("/creator", timeout.NewWithContext(mediaCon.GetCreatorByID, 10*time.Second))
	mediaRouter.Get("/genres/:kind", timeout.NewWithContext(mediaCon.GetGenres, 30*time.Second))
	// NOTE: singular to get a single genre, plural for more
	mediaRouter.Get("/genre/:kind/:genre", timeout.NewWithContext(mediaCon.GetGenre, 30*time.Second))
	mediaRouter.Post("/artists/by-name", timeout.NewWithContext(mediaCon.GetArtistsByName, 30*time.Second))
	mediaRouter.Post("/import", timeout.NewWithContext(mediaCon.ImportWeb, 60*time.Second))
}

func setupStatic(app *fiber.App, assets, artifacts string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	// stat both paths in parallel
	_, err := os.Stat(assets)
	if err != nil {
		return fmt.Errorf("failed to stat static assets directory. Ensure it's properly set and has the correct permissions: %v", err)
	}
	_, err = os.Stat(artifacts)
	if err != nil {
		return fmt.Errorf(
			`failed to stat frontend build artifacts directory.
			Ensure that you've run the JS bundler and it's properly set and has the correct permissions: %v`,
			err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		staticPath, err := filepath.Abs(artifacts)
		if err != nil {
			errChan <- fmt.Errorf("failed to get absolute path for static files: %w", err)
		}
		assetPath, err := filepath.Abs(assets)
		if err != nil {
			errChan <- fmt.Errorf("failed to get absolute path for static files: %w", err)
		}

		var mu sync.Mutex

		mu.Lock()
		app.Use("/static", filesystem.New(filesystem.Config{
			Root:         http.Dir(assetPath),
			Browse:       true,
			NotFoundFile: "404.html",
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
