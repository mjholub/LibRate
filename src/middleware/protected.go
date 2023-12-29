package middleware

import (
	"strings"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// Protected protect routes
func Protected(sess *session.Store, log *zerolog.Logger, conf *cfg.Config) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(conf.JWTSecret)},
		KeyFunc: func(token *jwt.Token) (interface{}, error) {
			return []byte(conf.JWTSecret), nil
		},
		ErrorHandler: jwtError,
		SuccessHandler: func(c *fiber.Ctx) error {
			tokenString := string(c.Request().Header.Peek("Authorization"))
			if log != nil {
				log.Debug().Msgf("JWT token: %s", strings.Split(tokenString, " ")[1])
			}
			s, err := sess.Get(c)
			if err != nil {
				return h.Res(c, fiber.StatusInternalServerError, "Failed to get session")
			}
			claims := jwt.MapClaims{
				"exp":         s.Get("claims_exp"),
				"member_name": s.Get("member_name"),
				"session_id":  s.ID(),
				"roles":       s.Get("claims_roles"),
			}
			token, err := jwt.ParseWithClaims(strings.Split(tokenString, " ")[1], claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(conf.JWTSecret), nil
			})
			if err != nil {
				if log != nil {
					log.Error().Err(err).Msgf("Failed to parse JWT: %s", err.Error())
				}
				return h.Res(c, fiber.StatusUnauthorized, "Invalid or expired JWT")
			}
			c.Locals("jwtToken", token)
			return c.Next()
		},
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	if err.Error() == "Missing or malformed JWT" {
		return h.ResData(c, fiber.StatusBadRequest, "Missing or malformed JWT", nil)
	}
	return h.ResData(c, fiber.StatusUnauthorized, "Invalid or expired JWT: "+err.Error(), nil)
}
