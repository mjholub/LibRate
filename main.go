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
	log := logging.Init()
	conf := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	if cfg.LoadConfig().IsError() {
		err := cfg.LoadConfig().Error()
		log.Warn().Msgf("failed to load config, using defaults: %v", err)
	}
	if DBRunning(conf.Port) {
		if *init {
			/* postgres 15 no longer supports pg_atoi
			pgUintCmd := exec.Command("./deps/pguint/make", "PG_CONFIG="+conf.PG_Config)
			if err := pgUintCmd.Run(); err != nil {
				panic(fmt.Errorf("failed to build pguint: %w", err))
			}
			if os.Getuid() != 0 {
				log.Warn().Msg("Not running as root, skipping database initialization (installing pguint requires root)")
				return
			}
			pgUintInstallCmd := exec.Command("./deps/pguint/make", "PG_CONFIG="+conf.PG_Config, "install")
			if err := pgUintInstallCmd.Run(); err != nil {
				panic(fmt.Errorf("failed to install pguint: %w", err))
			}
			log.Info().Msg("pguint installed")
			*/
			if err := db.InitDB(); err != nil {
				panic(err)
			}
			log.Info().Msg("Database initialized")
		}
	} else {
		log.Warn().Msgf("Database not running on port %d. Skipping initialization.", conf.Port)
	}
	app := fiber.New()
	app.Use(logger.New())

	// CORS
	app.Use("/api", func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", conf.Host+":"+strconv.Itoa(int(conf.Port)))
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		return c.Next()
	})

	routes.Setup(app)

	err := app.Listen(":3000")
	if err != nil {
		log.Panic().Err(err).Msg("Failed to listen on port 3000")
	}
}

func DBRunning(port uint16) bool {
	conn, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		return true // port in use => db running
	}
	conn.Close()
	return false
}
