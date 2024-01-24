package security

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"

	"codeberg.org/mjh/LibRate/cfg"
)

func SetupHelmet(conf *cfg.Config) (h fiber.Handler) {
	fh := conf.Fiber.Host
	fp := conf.Fiber.Port
	localAliases := strings.ReplaceAll(fmt.Sprintf(`%s:%d https://%s:%d http://%s:%d https://librate.localhost`,
		fh, fp, fh, fp, fh, fp), "'", "")
	return helmet.New(helmet.Config{
		XSSProtection:             "1; mode=block",
		ReferrerPolicy:            "no-referrer-when-downgrade",
		CrossOriginResourcePolicy: "cross-origin",
		CrossOriginEmbedderPolicy: "creadentialless",
		XFrameOptions:             "ALLOW FROM %s https://www.gravatar.com https://http.cat",
		ContentSecurityPolicy: fmt.Sprintf(`default-src 'self' https://gnu.org https://www.gravatar.com %s;
				style-src 'self' cdn.jsdelivr.net 'unsafe-inline';
				script-src 'self' https://unpkg.com/htmx.org@1.9.9 %s 'unsafe-inline' 'unsafe-eval';
				img-src 'self' * %s data: blob:;`,
			localAliases, localAliases, localAliases),
	})
}
