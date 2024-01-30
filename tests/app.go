// this file contains the seed code for a mockup test fiber app
package tests

import (
	"fmt"
	"net"
	"time"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func NewAppWithLogger(logger *zerolog.Logger) *fiber.App {
	fzlog := fiberzerolog.New(fiberzerolog.Config{
		Logger: logger,
	})

	app := fiber.New(
		fiber.Config{
			DisableStartupMessage: true,
			ReadTimeout:           60 * time.Second,
		},
	)
	app.Use(fzlog)

	return app
}

// first check if port 3100 is free, then loop until 65535 until a free port is found
func TryFindFreePort() (int, error) {
	for i := 3100; i < 65535; i++ {
		if listener, err := net.Listen("tcp", fmt.Sprintf(":%d", i)); err == nil {
			listener.Close()
			return i, nil
		}
	}
	return 0, fmt.Errorf("failed to find a free port for testing")
}
