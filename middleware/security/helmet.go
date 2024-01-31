package security

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"

	"codeberg.org/mjh/LibRate/cfg"
)

func SetupHelmet(conf *cfg.Config) (h fiber.Handler) {
	return helmet.New(helmet.Config{
		XSSProtection:             "1; mode=block",
		ReferrerPolicy:            "no-referrer-when-downgrade",
		CrossOriginResourcePolicy: "cross-origin",
		CrossOriginEmbedderPolicy: "creadentialless",
		XFrameOptions:             "ALLOW FROM %s https://www.gravatar.com https://http.cat",
		ContentSecurityPolicy: fmt.Sprintf(`default-src 'self' https://gnu.org https://www.gravatar.com %s;
				style-src 'self' cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css 'unsafe-inline';
				script-src 'self' %s 'unsafe-inline' 'unsafe-eval';
				img-src 'self' * %s data: blob:;`,
			conf.Fiber.Domain, conf.Fiber.Domain, conf.Fiber.Domain),
	})
}
