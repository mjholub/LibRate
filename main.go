package main

import (
	"flag"
	"net"
	"strconv"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/routes"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/idempotency"

	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/internal/logging"
)

func main() {
	init := flag.Bool("init", false, "Initialize database")
	flag.Parse()

	// TODO: get logging config from config file
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
	log := logging.Init(&logConf)

	// Load config
	conf, err := cfg.LoadConfig().Get()
	if err != nil {
		log.Panic().Msgf("failed to load config: %v", err)
	}
	if cfg.LoadConfig().IsError() {
		err = cfg.LoadConfig().Error()
		log.Warn().Msgf("failed to load config, using defaults: %v", err)
	}

	// database health check
	if DBRunning(conf.Port) {
		// Initialize database if it's running
		if *init {
			if err = db.InitDB(conf); err != nil {
				log.Panic().Err(err).Msg("Failed to initialize database")
			}
			log.Info().Msg("Database initialized")
		}
	} else {
		// FIXME: restore falling back to the default config if loading config fails
		// (bring back the cfg.ReadDefaults() function and put it in a call to .OrElse() error handler)
		log.Warn().
			Msgf("Database not running on port %d. Skipping initialization.", conf.Port)
	}

	// Connect to database
	dbConn, err := db.Connect(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer dbConn.Close()

	fiberlog := fiberzerolog.New(fiberzerolog.Config{
		Logger: &log,
		// skip logging for static files, there's too many of them
		SkipURIs: []string{"/_app/immutable", "/_app/chunks"},
	})
	// Create a new Fiber instance
	app := fiber.New()
	app.Use(fiberlog)
	app.Use(idempotency.New())

	// CORS
	setupCors(app)

	// Setup routes
	err = routes.Setup(&log, conf, dbConn, app, &fiberlog)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup routes")
	}

	// Listen on port 3000
	err = app.Listen(":3000")
	if err != nil {
		log.Panic().Err(err).Msg("Failed to listen on port 3000")
	}

	// Graceful shutdown
	err = app.ShutdownWithTimeout(time.Second * 10)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to shutdown app gracefully")
	}
}

func setupCors(app *fiber.App) {
	app.Use("/api", func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		return c.Next()
	})
}

func DBRunning(port uint16) bool {
	conn, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		return true // port in use => db running
	}
	conn.Close()
	return false
}
