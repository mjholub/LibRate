package main

import (
	"flag"
	"net"
	"strconv"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/internal/logging"
)

func main() {
	init := flag.Bool("init", false, "Initialize database")
	flag.Parse()
	// TODO: refactor so it looks better
	logConf := cfg.LoadLoggerConfig().OrElse(func(err error) logging.Config {
		return logging.Config{
			Level:  "info",
			Target: "stdout",
			Format: "json",
		}
	}(nil))
	log := logging.Init(&logConf)
	conf := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	if cfg.LoadConfig().IsError() {
		err := cfg.LoadConfig().Error()
		log.Warn().Msgf("failed to load config, using defaults: %v", err)
	}
	if DBRunning(conf.Port) {
		if *init {
			if err := db.InitDB(); err != nil {
				log.Panic().Err(err).Msg("Failed to initialize database")
			}
			log.Info().Msg("Database initialized")
		}
	} else {
		log.Warn().Msgf("Database not running on port %d. Skipping initialization.", conf.Port)
	}
	dbConn, err := db.Connect(&conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer dbConn.Close()
	app := fiber.New()
	app.Use(logger.New())

	// CORS
	setupCors(app)

	routes.Setup(&log, &conf, dbConn, app)

	err = app.Listen(":3000")
	if err != nil {
		log.Panic().Err(err).Msg("Failed to listen on port 3000")
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
