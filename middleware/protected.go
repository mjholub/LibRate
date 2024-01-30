package middleware

import (
	"fmt"
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
			s, err := sess.Get(c)
			if err != nil {
				return h.Res(c, fiber.StatusInternalServerError, "Failed to get session")
			}
			token, err := DecryptJWT(tokenString, s, conf)
			if err != nil {
				log.Error().Msgf("Failed to decrypt JWT: %v", err)
				return h.Res(c, fiber.StatusUnauthorized, "Invalid or expired JWT")
			}

			c.Locals("jwtToken", token)
			return c.Next()
		},
	})
}

func DecryptJWT(tokenString string, s *session.Session, conf *cfg.Config) (*jwt.Token, error) {
	claims := jwt.MapClaims{
		"exp":         s.Get("claims_exp"),
		"member_name": s.Get("member_name"),
		"webfinger":   s.Get("webfinger"),
		"session_id":  s.ID(),
		"roles":       s.Get("claims_roles"),
	}
	token, err := jwt.ParseWithClaims(
		strings.Split(
			tokenString, " ")[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(conf.JWTSecret), nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %v", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid or expired JWT")
	}
	return token, nil
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
