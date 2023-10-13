package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/routes"

	"github.com/avast/retry-go/v4"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jmoiron/sqlx"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/witer33/fiberpow"

	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/internal/logging"
)

func main() {
	init, NoDBSubprocess, ExternalDBHealthCheck, configFile, path, exit := parseFlags()
	// TODO: get logging config from config file
	log := initLogging()

	// Load config
	var (
		err  error
		conf *cfg.Config
	)
	if *configFile == "" {
		conf = cfg.LoadConfig().OrElse(&cfg.DefaultConfig)
	} else {
		conf, err = cfg.LoadFromFile(*configFile)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to load config file %s: %v", *configFile, err)
		}
	}

	// database first-run initialization
	// If the healtheck is to be handled externally, skip it
	dbRunning := DBRunning(*ExternalDBHealthCheck, conf.Port)
	dbConn, neo4jConn, err := connectDB(conf, *NoDBSubprocess)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to connect to database: %v", err)
	}
	log.Info().Msg("Connected to database")
	defer func() {
		if dbConn != nil {
			dbConn.Close()
		}
		if neo4jConn != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			neo4jConn.Close(ctx)
		}
	}()

	if *init {
		if !dbRunning {
			log.Warn().
				Msgf("Database not running on port %d. Not initializing.", conf.Port)
		}
		if err = initDB(conf, *NoDBSubprocess, *exit, &log); err != nil {
			log.Panic().Err(err).Msg("Failed to initialize database")
		}
	}

	if lo.Contains(os.Args, "migrate") {
		if err = db.Migrate(conf, *path); err != nil {
			log.Panic().Err(err).Msg("Failed to migrate database")
		}
		log.Info().Msg("Database migrated")
	}

	fiberlog := setupLogger(&log)

	// Create a new Fiber instance
	tag, err := getLatestTag()
	if err != nil {
		tag = "unknown"
	}
	app := fiber.New(fiber.Config{
		AppName:           fmt.Sprintf("LibRate %s", tag),
		Prefork:           conf.Fiber.Prefork,
		ReduceMemoryUsage: conf.Fiber.ReduceMemUsage,
	},
	)

	// proof of work based anti-spam/anti-ddos
	app.Use(recover.New())

	app.Use(fiberlog)

	app.Use(idempotency.New())
	// hardening
	app.Use(helmet.New())

	// setup secondary apps
	profilesApp, noscript, err := setupSecondaryApps(app, conf, fiberlog,
		recover.New(), idempotency.New(), helmet.New())
	if err != nil {
		log.Panic().Err(err).Msg("Failed to setup secondary apps")
	}
	apps := []*fiber.App{app, profilesApp, noscript}

	// CORS
	setupCors(apps)
	setupPOW(conf, apps)

	err = setupRoutes(conf, &log, dbConn, &neo4jConn, app, profilesApp, noscript, fiberlog)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to setup routes")
	}
	// Listen on port 3000
	err = modularListen(conf, app)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to listen")
	}
	// Graceful shutdown
	err = app.ShutdownWithTimeout(time.Second * 10)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to shutdown app gracefully")
	}
}

func setupPOW(conf *cfg.Config, app []*fiber.App) {
	for i := range app {
		app[i].Use(fiberpow.New(fiberpow.Config{
			PowInterval: time.Duration(conf.Fiber.PowInterval * int(time.Second)),
			Difficulty:  conf.Fiber.PowDifficulty,
			Filter: func(c *fiber.Ctx) bool {
				return c.IP() == conf.Fiber.Host || conf.LibrateEnv == "dev"
			},
			Storage: redis.New(redis.Config{
				Host:     conf.Redis.Host,
				Port:     conf.Redis.Port,
				Username: conf.Redis.Username,
				Password: conf.Redis.Password,
				Database: conf.Redis.Database,
			}),
		}))
	}
}

func setupLogger(logger *zerolog.Logger) fiber.Handler {
	fiberlog := fiberzerolog.New(fiberzerolog.Config{
		Logger: logger,
		// skip logging for static files, there's too many of them
		SkipURIs: []string{
			"/_app/immutable",
			"/_app/chunks",
			"/_app/*",
			"/profiles/_app",
			"/_app/immutable/chunks/",
		},
	})
	return fiberlog
}

func initDB(conf *cfg.Config, noSubprocess, exitAfter bool, logger *zerolog.Logger) error {
	// retry connecting to database
	err := retry.Do(
		func() error {
			return db.InitDB(conf, noSubprocess, exitAfter, logger)
		},
		retry.Attempts(5),
		retry.Delay(3*time.Second), // Delay between retries
		retry.OnRetry(func(n uint, err error) {
			logger.Info().Msgf("Attempt %d in initDB() failed: %v; retrying...", n, err)
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Info().Msg(`Database initialized.
			If you're seeing this when running from a container,
			necessary migrations will be run automatically.\n
			Otherwise you need to run them manually, either with -auto-migrate
			or by running "migrate -path db/migrations -database <your db connection string>\n
			(use -exit to exit after migrations are run)"
			`)
	return nil
}

func setupCors(apps []*fiber.App) {
	for i := range apps {
		apps[i].Use("/api", func(c *fiber.Ctx) error {
			c.Set("Access-Control-Allow-Origin", "*")
			c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			return c.Next()
		})
	}
}

func DBRunning(skipCheck bool, port uint16) bool {
	if skipCheck {
		return true
	}
	conn, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		return true // port in use => db running
	}
	conn.Close()
	return false
}

func parseFlags() (*bool, *bool, *bool, *string, *string, *bool) {
	init := flag.Bool("init", false, "Initialize database")
	NoDBSubprocess := flag.Bool("no-db-subprocess", false,
		"Do not launching database as subprocess if not running. Not recommended in containers.")
	ExternalDBHealthCheck := flag.Bool("hc-extern", false,
		`Skips calling the built-in database health check. 
		Useful for containers with external databases, where pg_isready is used instead.`)
	configFile := flag.String("config", "config.yml", "Path to config file")
	path := flag.String("path", "db/migrations", "Path to migrations")
	exit := flag.Bool("exit", false, "Exit after running migrations")
	flag.Parse()

	return init, NoDBSubprocess, ExternalDBHealthCheck, configFile, path, exit
}

func initLogging() zerolog.Logger {
	logConf := logging.Config{
		Level:  "debug",
		Target: "stdout",
		Format: "json",
		Caller: true,
		Timestamp: logging.TimestampConfig{
			Enabled: true,
			Format:  "2006-01-02T15:04:05.000Z07:00",
		},
	}

	return logging.Init(&logConf)
}

func connectDB(conf *cfg.Config, noSubprocess bool) (*sqlx.DB, neo4j.DriverWithContext, error) {
	// Connect to database
	var (
		err       error
		dbConn    *sqlx.DB
		neo4jConn neo4j.DriverWithContext
	)
	switch conf.Engine {
	case "postgres", "mariadb", "sqlite":
		dbConn, err = db.Connect(conf, noSubprocess)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		return dbConn, nil, nil
	case "neo4j":
		neo4jConn, err = db.ConnectNeo4j(conf)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to neo4j: %w", err)
		}
		return nil, neo4jConn, nil
	default:
		return nil, nil, fmt.Errorf("unsupported database engine \"%q\" or error reading config", conf.Engine)
	}
}

// unsure if middlewares need to be re-allocated for each subapp
func setupSecondaryApps(mainApp *fiber.App, conf *cfg.Config, middlewares ...interface{}) (*fiber.App, *fiber.App, error) {
	profilesApp := fiber.New()
	profilesApp.Static("/", "./fe/build/")
	profilesApp.Use(middlewares...)
	mainApp.Mount("/profiles", profilesApp)
	noscript, err := setupNoscript()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup noscript app: %w", err)
	}
	noscript.Use(middlewares...)
	mainApp.Mount("/noscript", noscript)
	return profilesApp, noscript, nil
}

func modularListen(conf *cfg.Config, app *fiber.App) error {
	listenPort := strconv.Itoa(conf.Fiber.Port)
	if listenPort == "" {
		listenPort = "3000"
	}
	listenHost := conf.Fiber.Host
	if listenHost == "" {
		listenHost = "127.0.0.1"
	}
	listenAddr := fmt.Sprintf("%s:%s", listenHost, listenPort)
	err := app.Listen(listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}
	return nil
}

func setupRoutes(
	conf *cfg.Config,
	log *zerolog.Logger,
	dbConn *sqlx.DB,
	neo4jConn *neo4j.DriverWithContext,
	app, profilesApp, noscript *fiber.App,
	fiberlog fiber.Handler,
) (err error) {
	err = routes.SetupProfiles(log, conf, dbConn, neo4jConn, profilesApp, &fiberlog)
	if err != nil {
		return fmt.Errorf("failed to setup profiles routes: %w", err)
	}
	switch conf.Engine {
	case "postgres", "sqlite", "mariadb":
		err = routeNoScript(noscript, dbConn, log, conf, nil)
	case "neo4j":
		err = routeNoScript(noscript, nil, log, conf, *neo4jConn)
	default:
		return fmt.Errorf("unsupported database engine \"%q\" or error reading config", conf.Engine)
	}
	if err != nil {
		return fmt.Errorf("failed to setup noscript routes: %w", err)
	}

	// Setup routes
	err = routes.Setup(log, conf, dbConn, neo4jConn, app, &fiberlog)
	if err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}
	return nil
}
