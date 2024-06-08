// Copyright (C) 2023-2024 LibRate contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/cmd"
	"codeberg.org/mjh/LibRate/lib/redist"
	"codeberg.org/mjh/LibRate/routes"

	"github.com/go-playground/validator/v10"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/witer33/fiberpow"

	_ "codeberg.org/mjh/LibRate/static/meta" // swagger docs

	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/internal/logging"
	"codeberg.org/mjh/LibRate/middleware/profiling"
	"codeberg.org/mjh/LibRate/middleware/render"
	"codeberg.org/mjh/LibRate/middleware/session"
)

type FlagArgs struct {
	// init is a flag to initialize the database
	Init bool
	// configFile is a flag to specify the path to the config file
	ConfigFile string
	// Whether to start the profiling/tracing server
	// Due to security reasons, this is only available in development mode
	Profile bool
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
		pgConn             *pgxpool.Pool
		err                error
		conf               *cfg.Config
		validationProvider = validator.New()
	)

	if flags.ConfigFile == "" {
		conf = cfg.LoadConfig().OrElse(&cfg.DefaultConfig)
	} else {
		conf, err = cfg.LoadFromFile(flags.ConfigFile)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to load config file %s: %v", flags.ConfigFile, err)
		}
	}

	if conf.LibrateEnv == "development" && flags.Profile {
		go func() {
			profiling.Serve(&log)
		}()
	}

	log = initLogging(&conf.Logging)
	log.Info().Msgf("Reloaded logger with the custom config: %+v", conf.Logging)
	validationErrors := cfg.Validate(conf, validationProvider)
	if len(validationErrors) > 0 {
		for i := range validationErrors {
			log.Warn().Msgf("Validation error: %+v", validationErrors[i])
		}
		log.Fatal().Msg("errors were encountered while validating the config. Exiting.")
	}
	log.Debug().Msgf("Config: %+v", conf)

	searchCache := redis.New(redis.Config{
		Host:     conf.Redis.Host,
		Port:     conf.Redis.Port,
		Username: conf.Redis.Username,
		Password: conf.Redis.Password,
		Database: conf.Redis.SearchDB,
	})

	// Create a new Fiber instance
	app := cmd.CreateApp(conf)
	s := &cmd.GrpcServer{
		App:    app,
		Log:    &log,
		Config: &conf.GRPC,
	}
	go cmd.RunGrpcServer(s)

	staticDirAbs, err := filepath.Abs(conf.Fiber.StaticDir)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get absolute path of static directory %s", conf.Fiber.StaticDir)
	}

	// Setup templated pages, like privacy policy and TOS
	pagesCache := render.SetupCaching(conf)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err = render.
			WatchFiles(ctx, staticDirAbs, &log, pagesCache); err != nil {
			log.Error().Err(err).Msg("Error watching files")
		}
	}()

	// database first-run initialization
	pgConn, err = connectDB(conf)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to connect to database: %v", err)
	}
	log.Info().Msg("Connected to database")
	defer func() {
		if pgConn != nil {
			pgConn.Close()
		}
	}()

	if flags.Init {
		if err = initDB(&conf.DBConfig, flags.Init, &log); err != nil {
			log.Panic().Err(err).Msg("Failed to initialize database")
		}
	}

	// Check password entropy
	entropy, _ := redist.CheckPasswordEntropy(conf.Secret)
	if err == nil && entropy < 50 {
		log.Warn().Msgf("Secret is weak: %2f bits of entropy", entropy)
	}

	// Setup session
	sess, err := session.Setup(conf)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to setup session")
	}
	log.Info().Msg("Session handler ready")

	// Setup Proof of Work antispam middleware
	setupPOW(conf, app)

	// setup other middlewares
	// By default, the app uses the following:
	// cache, cors, csrf, helmet, recover, idempotency, etag, compress
	var wg sync.WaitGroup
	middlewares := cmd.SetupMiddlewares(conf, &log)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range middlewares {
			app.Use(middlewares[i])
			if i == len(middlewares)-1 {
				log.Info().Msg("Middlewares set up")
			}
		}
	}()

	// setup logging
	fzlog := cmd.SetupLogger(conf, &log)
	app.Use(fzlog)
	log.Info().Msg("Logger set up")

	// set up websocket
	wsConfig := cmd.SetupWS(app, "/search")
	log.Info().Msg("Websocket set up")
	wg.Wait()

	render.SetupTemplatedPages(
		conf.Fiber.DefaultLanguage,
		app, &log, pagesCache)

	log.Info().Msg("Templated pages set up")

	r := routes.RouterProps{
		Conf:            conf,
		Log:             &log,
		LogHandler:      fzlog,
		DB:              pgConn,
		App:             app,
		SessionHandler:  sess,
		WebsocketConfig: &wsConfig,
		Validation:      validationProvider,
		Cache:           searchCache,
	}

	log.Info().Msg("Setting up routes")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err = setupRoutes(ctx, &r)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to setup routes")
	}
	log.Info().Msg("Routes set up")

	// Listen on chosen port, host and protocol
	// (disabling HTTPS still works if you use reverse proxy)
	err = modularListen(conf, app)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to listen")
	}
}

// Set up a Proof of Work middleware
// This is an ethical alternative to things like Cloudflare
// Basically, an attacker would have to run the same computation as the server
// (granted the difficulty, which is measured in
// the number of calculations required to find a SHA256 hash with a certain number of leading zeroes,
// is set high enough, and the check frequency is set low enough)
// to access a resource or perform an action
// One thing to keep in mind when setting the difficulty and
// check frequency is that a too high difficulty with too frequent checks
// might significantly slow down the page,
// harm SEO and drain battery on mobile devices
// A good value for stack where you have additional measures in place
// like a rate limiter on your reverse proxy
// lies somewhere between 15-45kH/s difficulty and 5-15 minutes check frequency
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

func initDB(dbConf *cfg.DBConfig, do bool, logger *zerolog.Logger) error {
	if !do {
		return nil
	}
	dbRunning := DBRunning(dbConf.Port)
	if dbRunning {
		logger.Warn().Msgf("Database already running on port %d. Not initializing.", dbConf.Port)
		return nil
	}

	err := db.InitDB(dbConf, logger)
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
		init, profiler         bool
		configFile, skipErrors string
	)

	const (
		initVal    = false
		initUse    = "Initialize database"
		confVal    = "config.yml"
		confUse    = "Path to config file"
		pprofVal   = false
		pprofUse   = "Start tracing/profiling server. LibrateEnv must be set to development"
		skipErrVal = ""
		skipErrUse = "Comma-separated list of error codes to skip and not panic on"
		short      = " (shorthand)"
	)

	flag.BoolVar(&init, "init", initVal, initUse)
	flag.BoolVar(&init, "i", initVal, initUse+short)
	flag.StringVar(&configFile, "config", confVal, confUse)
	flag.StringVar(&configFile, "c", confVal, confUse+short)
	flag.StringVar(&skipErrors, "skip-errors", skipErrVal, skipErrUse)
	flag.StringVar(&skipErrors, "s", skipErrVal, skipErrUse+short)
	flag.BoolVar(&profiler, "tracing", pprofVal, pprofUse)
	flag.BoolVar(&profiler, "t", pprofVal, pprofUse+short)

	flag.Parse()

	return FlagArgs{
		Init:       init,
		ConfigFile: configFile,
		Profile:    profiler,
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

func connectDB(conf *cfg.Config) (*pgxpool.Pool, error) {
	dsn := db.CreateDsn(&conf.DBConfig)

	return db.Connect(context.Background(), dsn, conf.RetryAttempts)
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

func setupRoutes(ctx context.Context, r *routes.RouterProps) (err error) {
	// Setup routes
	err = routes.Setup(ctx, r)
	if err != nil {
		return fmt.Errorf("failed to setup routes: %v", err)
	}
	return nil
}
