package auth

import (
	"fmt"

	"codeberg.org/mjh/internal/crypt"

	"codeberg.org/mjh/LibRate/internal/crypt"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/redis/v3"
)

// GetPubKey returns a public key for the client to encrypt their password
// @Summary Get public key
// @Description Get public key for client to encrypt their password
// @Tags Auth
// @Produce json
// @Success 200 {object} PublicKeyResponse
// @Router /auth/pubkey [get]
// @Security ApiKeyAuth
// @Security OAuth2Application[write]
func (a *Service) GetPubKey(c *fiber.Ctx) error {
	keys, err := crypt.GenerateKeys()
	if err != nil {
		return err
	}

	// as redis does not support multi-value assignments,
	// we need first to convert the keys struct into a json string that
	// we can then deserialize upon retrieval
	var jsonString []byte
	jsonString, err = json.Marshal(keys)
	if err != nil {
		return fmt.Errorf("could not marshal keys: %w", err)
	}

	// TODO: extend the Service struct
	expiration := a.config.Auth.PubKeyExpiration

	// store the keys in redis
	err = a.redis.Set(keys.Identifier, jsonString, expiration)
	if err != nil {
		return fmt.Errorf("could not store keys in redis: %w", err)
	}

	return nil
}
