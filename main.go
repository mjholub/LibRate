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
	"codeberg.org/mjh/LibRate/cmd"
	"codeberg.org/mjh/LibRate/lib/redist"
	"codeberg.org/mjh/LibRate/routes"

	"github.com/avast/retry-go/v4"
	"github.com/gofiber/fiber/v2"
	fiberSession "github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jmoiron/sqlx"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/witer33/fiberpow"

	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/internal/crypt"
	"codeberg.org/mjh/LibRate/internal/logging"
	"codeberg.org/mjh/LibRate/middleware/session"
)

type FlagArgs struct {
	// init is a flag to initialize the database
	Init bool
	// NoDBSubprocess is a flag to not launch the database as a subprocess if it is not running
	NoDBSubprocess bool
	// ExternalDBHealthCheck is a flag to skip the built-in healthcheck, especially for database
	// Should be used in containers with external databases, where pg_isready is used instead
	ExternalDBHealthCheck bool
	// configFile is a flag to specify the path to the config file
	ConfigFile string
	// path is a flag to specify the path to the migrations that should be applied.
	// TODO: add this feature (currently only batch application of all migrations is supported)
	Path string
	// When exit is true, the program will exit after running migrations
	Exit bool
	// SkipErrors is a comma-separated list of error codes to skip and not panic on.
	// Particularly useful in development to bypass certain less important blockers
	SkipErrors string
}

func main() {
	flags := parseFlags()
	// first, start logging with some opinionated defaults, just for the config loading phase
	log := initLogging(nil)

	log.Info().Msg("Starting LibRate")
	// Load config
	var (
		err  error
		conf *cfg.Config
	)

	if flags.ConfigFile == "" {
		conf = cfg.LoadConfig().OrElse(&cfg.DefaultConfig)
	} else {
		conf, err = cfg.LoadFromFile(flags.ConfigFile)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to load config file %s: %v", flags.ConfigFile, err)
		}
	}
	log = initLogging(&conf.Logging)
	log.Info().Msgf("Reloaded logger with the custom config: %+v", conf.Logging)

	// database first-run initialization
	// If the healtheck is to be handled externally, skip it
	dbConn, neo4jConn, err := connectDB(conf, flags.NoDBSubprocess)
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

	if err = initDB(conf, flags.Init, flags.ExternalDBHealthCheck, flags.NoDBSubprocess, flags.Exit, &log); err != nil {
		log.Panic().Err(err).Msg("Failed to initialize database")
	}

	err = handleMigrations(conf, &log, flags.Path)
	if err != nil {
		log.Panic().Err(err).Msg(err.Error())
	}

	entropy, _ := redist.CheckPasswordEntropy(conf.Secret)
	if err == nil && entropy < 50 {
		log.Warn().Msgf("Secret is weak: %2f bits of entropy", entropy)
	}

	sessionsDB, err := crypt.CreateFile("librate_sessions.db")
	if err != nil {
		log.Panic().Err(err).Msg("Failed to create session files")
	}
	defer sessionsDB.Close()

	cryptStore, err := crypt.CreateStorage("librate_sessions.db", conf.Secret)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to create session storage")
	}

	// Setup session
	sess, err := session.Setup(conf, cryptStore)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to setup session")
	}

	// Create a new Fiber instance
	app := cmd.CreateApp(conf)
	middlewares := cmd.SetupMiddlewares(conf, &log, sess)
	go func() {
		for i := range middlewares {
			app.Use(middlewares[i])
		}
	}()

	// setup secondary apps
	profilesApp, err := setupSecondaryApps(app, middlewares)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to setup secondary apps")
	}
	apps := []*fiber.App{app, profilesApp}

	setupPOW(conf, apps)

	err = setupRoutes(conf, &log, dbConn, &neo4jConn, app,
		profilesApp, sess)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to setup routes")
	}

	// Listen on chosen port, host and protocol
	// (disabling HTTPS still works if you use reverse proxy)
	err = modularListen(conf, app)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to listen")
	}

	// Graceful or timed shutdown
	if conf.Fiber.ShutdownTimeout >= 0 {
		err = app.ShutdownWithTimeout(time.Second * time.Duration(conf.Fiber.ShutdownTimeout))
		if err != nil {
			log.Panic().Err(err).Msgf("Failed to shutdown app: %v", err)
		}
	} else {
		err = app.Shutdown()
		if err != nil {
			log.Panic().Err(err).Msgf("Failed to shutdown app: %v", err)
		}
	}
}

func setupPOW(conf *cfg.Config, app []*fiber.App) {
	if conf.Fiber.PowDifficulty == 0 {
		conf.Fiber.PowDifficulty = 60000
	}
	for i := range app {
		app[i].Use(fiberpow.New(fiberpow.Config{
			PowInterval: time.Duration(conf.Fiber.PowInterval * int(time.Second)),
			Difficulty:  conf.Fiber.PowDifficulty,
			Filter: func(c *fiber.Ctx) bool {
				return c.IP() == conf.Fiber.Host || conf.LibrateEnv == "development"
			},
			Storage: redis.New(redis.Config{
				Host:     conf.Redis.Host,
				Port:     conf.Redis.Port,
				Username: conf.Redis.Username,
				Password: conf.Redis.Password,
				Database: conf.Redis.PowDB,
			}),
		}))
	}
}

func initDB(conf *cfg.Config, do, externalHC, noSubprocess, exitAfter bool, logger *zerolog.Logger) error {
	if !do {
		return nil
	}
	dbRunning := DBRunning(externalHC, conf.Port)
	if dbRunning {
		logger.Warn().Msgf("Database already running on port %d. Not initializing.", conf.Port)
		return nil
	}
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

func parseFlags() FlagArgs {
	var (
		init, NoDBSubprocess, ExternalDBHealthCheck, exit bool
		configFile, path, skipErrors                      string
	)

	const (
		initVal         = false
		initUse         = "Initialize database"
		noDBSPVal       = false
		noDBSPUse       = `Skip launching the database as subprocess if not running.`
		externalDBHCVal = false
		exDBHCUse       = `Skip calling the built-in database health check.`
		confVal         = "config.yml"
		confUse         = "Path to config file"
		skipErrVal      = ""
		skipErrUse      = "Comma-separated list of error codes to skip and not panic on"
		pathVal         = "db/migrations"
		pathUse         = "Path to migrations to apply"
		exitVal         = false
		exitUse         = "Exit after running migrations"
	)
	flag.BoolVar(&init, "init", initVal, initUse)
	flag.BoolVar(&init, "i", initVal, initUse+" (&shorthand)")
	flag.BoolVar(&NoDBSubprocess, "no-db-subprocess", noDBSPVal, noDBSPUse)
	flag.BoolVar(&NoDBSubprocess, "n", noDBSPVal, noDBSPUse+" (&shorthand)")
	flag.BoolVar(&ExternalDBHealthCheck, "external-db-health-check", externalDBHCVal, exDBHCUse)
	flag.BoolVar(&ExternalDBHealthCheck, "e", externalDBHCVal, exDBHCUse+" (&shorthand)")
	flag.StringVar(&configFile, "config", confVal, confUse)
	flag.StringVar(&configFile, "c", confVal, confUse+" (&shorthand)")
	flag.StringVar(&skipErrors, "skip-errors", skipErrVal, skipErrUse)
	flag.StringVar(&skipErrors, "s", skipErrVal, skipErrUse+" (&shorthand)")
	flag.StringVar(&path, "path", pathVal, pathUse)
	flag.StringVar(&path, "p", pathVal, pathUse+" (&shorthand)")
	flag.BoolVar(&exit, "exit", exitVal, exitUse)
	flag.BoolVar(&exit, "x", exitVal, exitUse+" (&shorthand)")

	flag.Parse()

	return FlagArgs{
		Init:                  init,
		NoDBSubprocess:        NoDBSubprocess,
		ExternalDBHealthCheck: ExternalDBHealthCheck,
		ConfigFile:            configFile,
		Path:                  path,
		Exit:                  exit,
		SkipErrors:            skipErrors,
	}
}

func initLogging(logConf *logging.Config) zerolog.Logger {
	if logConf == nil {
		logConf = &logging.Config{
			Level:  "trace",
			Target: "stdout",
			Format: "json",
			Caller: true,
			Timestamp: logging.TimestampConfig{
				Enabled: true,
				Format:  "2006-01-02T15:04:05.000Z07:00",
			},
		}

		return logging.Init(logConf)
	}
	return logging.Init(logConf)
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
func setupSecondaryApps(mainApp *fiber.App,
	middlewares []fiber.Handler,
) (*fiber.App, error) {
	profilesApp := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
	})
	profilesApp.Static("/", "./fe/build/")
	for i := range middlewares {
		profilesApp.Use(middlewares[i])
	}
	mainApp.Mount("/profiles", profilesApp)
	return profilesApp, nil
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
	if conf.Fiber.TLS {
		err := app.ListenTLS(listenAddr, conf.Keys.Public, conf.Keys.Private)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
		}
	}
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
	app, profilesApp *fiber.App,
	sess *fiberSession.Store,
) (err error) {
	err = routes.SetupProfiles(log, conf, dbConn, neo4jConn, profilesApp)
	if err != nil {
		return fmt.Errorf("failed to setup profiles routes: %w", err)
	}

	// Setup routes
	err = routes.Setup(log, conf, dbConn, neo4jConn, app, sess)
	if err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}
	return nil
}

func handleMigrations(conf *cfg.Config, log *zerolog.Logger, path string) error {
	if !lo.Contains(os.Args, "migrate") {
		return nil
	}

	if err := db.Migrate(conf, path); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Info().Msg("Database migrated")
	return nil
}
