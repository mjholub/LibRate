package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/trace"
	"strconv"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/cmd"
	"codeberg.org/mjh/LibRate/lib/redist"
	"codeberg.org/mjh/LibRate/routes"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	fiberSession "github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/witer33/fiberpow"

	_ "net/http/pprof"

	_ "codeberg.org/mjh/LibRate/static/meta" // swagger docs

	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/internal/logging"
	"codeberg.org/mjh/LibRate/middleware/render"
	"codeberg.org/mjh/LibRate/middleware/session"
)

type FlagArgs struct {
	// init is a flag to initialize the database
	Init bool
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

// @title LibRate
// @version dev
// @description API for LibRate, a social media cataloguing and reviewing service

// @contact.name MJH
// @contact.email TODO@flagship.instance

// @license.name GNU Affero General Public License v3
// @license.url https://www.gnu.org/licenses/agpl-3.0.html

// @BasePath /api
func main() {
	flags := parseFlags()
	// first, start logging with some opinionated defaults, just for the config loading phase
	log := initLogging(nil)

	log.Info().Msg("Starting LibRate")
	// Load config
	var (
		dbConn    *sqlx.DB
		pgConn    *pgxpool.Pool
		err       error
		conf      *cfg.Config
		validator = validator.New()
	)

	if flags.ConfigFile == "" {
		conf = cfg.LoadConfig().OrElse(&cfg.DefaultConfig)
	} else {
		conf, err = cfg.LoadFromFile(flags.ConfigFile)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to load config file %s: %v", flags.ConfigFile, err)
		}
	}

	if conf.LibrateEnv == "development" {
		go func() {
			log.Info().Msg("Starting pprof server")
			err = http.ListenAndServe("localhost:6060", nil)
			if err != nil {
				log.Panic().Err(err).Msg("Failed to start pprof server")
			}
			f, err := os.Create("trace.out")
			if err != nil {
				panic(err)
			}
			defer f.Close()
			if err := trace.Start(f); err != nil {
				log.Panic().Err(err).Msg("Failed to start trace")
			}
			defer trace.Stop()
		}()
	}

	log = initLogging(&conf.Logging)
	log.Info().Msgf("Reloaded logger with the custom config: %+v", conf.Logging)
	validationErrors := cfg.Validate(conf, validator)
	if len(validationErrors) > 0 {
		for i := range validationErrors {
			log.Warn().Msgf("Validation error: %+v", validationErrors[i])
		}
		log.Fatal().Msg("errors were encountered while validating the config. Exiting.")
	}
	log.Debug().Msgf("Config: %+v", conf)

	// Create a new Fiber instance
	app := cmd.CreateApp(conf)
	s := &cmd.GrpcServer{
		App:    app,
		Log:    &log,
		Config: &conf.GRPC,
	}
	go cmd.RunGrpcServer(s)

	// Setup templated pages, like privacy policy and TOS
	go func() {
		pages, err := render.MarkdownToHTML(conf.Fiber.StaticDir)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to render pages from markdown")
		}
		for i := range pages {
			app.Get("/"+pages[i].Name, func(c *fiber.Ctx) error {
				c.Set("Content-Type", "text/html")
				return c.Send(pages[i].Data)
			})
		}
	}()

	// database first-run initialization
	dbConn, pgConn, err = connectDB(conf)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to connect to database: %v", err)
	}
	log.Info().Msg("Connected to database")
	defer func() {
		if dbConn != nil {
			dbConn.Close()
		}
		if pgConn != nil {
			pgConn.Close()
		}
	}()

	if flags.Init {
		if err = initDB(&conf.DBConfig, flags.Init, flags.Exit, &log); err != nil {
			log.Panic().Err(err).Msg("Failed to initialize database")
		}
	}

	err = handleMigrations(conf, &log, flags.Path)
	if err != nil {
		log.Panic().Err(err).Msg(err.Error())
	}

	entropy, _ := redist.CheckPasswordEntropy(conf.Secret)
	if err == nil && entropy < 50 {
		log.Warn().Msgf("Secret is weak: %2f bits of entropy", entropy)
	}

	// Setup session
	sess, err := session.Setup(conf)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to setup session")
	}

	middlewares := cmd.SetupMiddlewares(conf, &log)
	go func() {
		for i := range middlewares {
			app.Use(middlewares[i])
		}
	}()
	fzlog := cmd.SetupLogger(conf, &log)
	app.Use(fzlog)

	setupPOW(conf, app)

	wsConfig := cmd.SetupWS(app, "/search")
	err = setupRoutes(conf, &log, fzlog, pgConn, dbConn, app, sess, wsConfig)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to setup routes")
	}

	// Listen on chosen port, host and protocol
	// (disabling HTTPS still works if you use reverse proxy)
	err = modularListen(conf, app)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to listen")
	}
}

func setupPOW(conf *cfg.Config, app *fiber.App) {
	if conf.Fiber.PowDifficulty == 0 {
		conf.Fiber.PowDifficulty = 60000
	}
	app.Use(fiberpow.New(fiberpow.Config{
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

func initDB(dbConf *cfg.DBConfig, do, exitAfter bool, logger *zerolog.Logger) error {
	if !do {
		return nil
	}
	dbRunning := DBRunning(dbConf.Port)
	if dbRunning {
		logger.Warn().Msgf("Database already running on port %d. Not initializing.", dbConf.Port)
		return nil
	}

	err := db.InitDB(dbConf, exitAfter, logger)
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

func DBRunning(port uint16) bool {
	conn, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		return true // port in use => db running
	}
	conn.Close()
	return false
}

func parseFlags() FlagArgs {
	var (
		init, exit                   bool
		configFile, path, skipErrors string
	)

	const (
		initVal    = false
		initUse    = "Initialize database"
		confVal    = "config.yml"
		confUse    = "Path to config file"
		skipErrVal = ""
		skipErrUse = "Comma-separated list of error codes to skip and not panic on"
		pathVal    = "db/migrations"
		pathUse    = "Path to migrations to apply"
		exitVal    = false
		exitUse    = "Exit after running migrations"
		short      = " (shorthand)"
	)
	flag.BoolVar(&init, "init", initVal, initUse)
	flag.BoolVar(&init, "i", initVal, initUse+short)
	flag.StringVar(&configFile, "config", confVal, confUse)
	flag.StringVar(&configFile, "c", confVal, confUse+short)
	flag.StringVar(&skipErrors, "skip-errors", skipErrVal, skipErrUse)
	flag.StringVar(&skipErrors, "s", skipErrVal, skipErrUse+short)
	flag.StringVar(&path, "path", pathVal, pathUse)
	flag.StringVar(&path, "p", pathVal, pathUse+short)
	flag.BoolVar(&exit, "exit", exitVal, exitUse)
	flag.BoolVar(&exit, "x", exitVal, exitUse+short)

	flag.Parse()

	return FlagArgs{
		Init:       init,
		ConfigFile: configFile,
		Path:       path,
		Exit:       exit,
		SkipErrors: skipErrors,
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

func connectDB(conf *cfg.Config) (*sqlx.DB, *pgxpool.Pool, error) {
	// Connect to database
	var (
		err    error
		dbConn *sqlx.DB
	)

	dsn := db.CreateDsn(&conf.DBConfig)

	switch conf.Engine {
	// case "postgres", "mariadb", "sqlite":
	case "postgres":
		dbConn, err = db.Connect(conf.Engine, dsn, conf.RetryAttempts)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		pgConnPool, err := pgxpool.New(context.Background(), dsn)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		return dbConn, pgConnPool, nil
	default:
		return nil, nil, fmt.Errorf("unsupported database engine \"%q\" or error reading config", conf.Engine)
	}
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
	fzlog fiber.Handler,
	pgConn *pgxpool.Pool,
	dbConn *sqlx.DB,
	app *fiber.App,
	sess *fiberSession.Store,
	wsConfig websocket.Config,
) (err error) {
	// Setup routes
	err = routes.Setup(log, fzlog, conf, dbConn, pgConn, app, sess, wsConfig)
	if err != nil {
		return fmt.Errorf("failed to setup routes: %v", err)
	}
	return nil
}

func handleMigrations(conf *cfg.Config, log *zerolog.Logger, path string) error {
	if !lo.Contains(os.Args, "migrate") {
		return nil
	}

	if err := db.Migrate(log, conf, path); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Info().Msg("Database migrated")
	return nil
}
