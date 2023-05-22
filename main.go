package main

import (
	"net/http"
	"path/filepath"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

func main() {
	log := utils.NewLogger()
	log.Info("Starting server...")
	app := fiber.New()
	cfg.LoadConfig()

	staticPath, err := filepath.Abs("./fe/public")
	if err != nil {
		log.Sugar().Fatalf("Error loading static path: %s", err)
	}

	app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.Dir(staticPath),
		Browse: true,
	}))
	app.Listen(":3000")
}
