package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/routes"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/recover"

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
		log.Warn().Msgf("failed to load config, using defaults: %v", err)
		conf = &cfg.DefaultConfig
	}

	// database first-run initialization
	if DBRunning(conf.Port) {
		if *init {
			if err = db.InitDB(conf); err != nil {
				log.Panic().Err(err).Msg("Failed to initialize database")
			}
			log.Info().Msg("Database initialized")
		}
	} else {
		log.Warn().
			Msgf("Database not running on port %d. Skipping initialization.", conf.Port)
	}

	// Connect to database
	dbConn, err := db.Connect(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	log.Info().Msg("Connected to database")
	defer dbConn.Close()

	fiberlog := fiberzerolog.New(fiberzerolog.Config{
		Logger: &log,
		// skip logging for static files, there's too many of them
		SkipURIs: []string{
			"/_app/immutable",
			"/_app/chunks",
			"/profiles/_app",
			"/_app/immutable/chunks/",
		},
	})
	// Create a new Fiber instance
	app := fiber.New()
	app.Use(recover.New())

	profilesApp := fiber.New()
	profilesApp.Static("/", "./fe/build/")
	app.Mount("/profiles", profilesApp)
	profilesApp.Use(fiberlog)
	// redirect GET requests to /profiles/_app one directory up
	err = routes.SetupProfiles(&log, conf, dbConn, profilesApp, &fiberlog)
	if err != nil {
		log.Error().Err(err).Msg("Failed to setup profiles routes")
	}
	// fallback using a templating engine in case sveltekit breaks
	noscript, err := setupNoscript()
	if err != nil {
		log.Error().Err(err).Msg("Failed to setup noscript app")
	}
	app.Mount("/noscript", noscript)
	err = routeNoScript(noscript, dbConn, &log, conf)
	if err != nil {
		log.Error().Err(err).Msg("Failed to setup noscript routes")
	}

	app.Use(fiberlog)
	app.Use(idempotency.New())

	// CORS
	setupCors(app, conf)

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

func setupCors(app *fiber.App, conf *cfg.Config) {
	app.Use("/api", func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", fmt.Sprintf("http://%s:%d", conf.Host, conf.Port))
		c.Set("Access-Control-Allow-Origin", fmt.Sprintf("http://localhost:%d", conf.Port))
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
