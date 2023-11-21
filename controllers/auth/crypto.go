package auth

import (
	"fmt"

	"filippo.io/age"

	"codeberg.org/mjh/LibRate/internal/crypt"
	h "codeberg.org/mjh/LibRate/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
)

type keyData struct {
	// key itself
	identifier       string            `json:""`
	privPubNestedMap map[string]string `json:""`
}

// GetPubKey returns a public key and uuid for the client to encrypt their password
// In the context of X25519 it returns the recipient's public key
// @Summary Get public key
// @Description Get public key for client to encrypt their password
// @Tags Auth
// @Produce json
// @Success 200 {object} PublicKeyResponse
// @Router /auth/pubkey [get]
// @Security ApiKeyAuth
// @Security OAuth2Application[write]
func (a *Service) GetPubKey(c *fiber.Ctx) error {
	// get the X25519 keypair and the associated UUID
	keys, err := crypt.GenerateKeys()
	if err != nil {
		return err
	}

	idStr := keys.Identifier.String()
	if idStr == "00000000-0000-0000-0000-000000000000" {
		a.log.Error().Msg("error parsing key identifier to string: zeroed UUID")
		return h.Res(c, fiber.StatusInternalServerError, "Internal server error")
	}

	const (
		private = "private"
		public  = "public"
	)

	// save the data to secrets storage to be retrieved later
	// pubKey is the key seeded upon application start, used to encrypt the temporary secrets
	// storage

	_, err = a.secStorage.Exec(
		fmt.Sprintf(`INSERT INTO keys(id, private, public) VALUES('%s', '%s', '%s')`,
			idStr, keys.Private.String(), keys.Public.String()))
	if err != nil {
		a.log.Error().Msgf("error inserting key data into secrets storage: %v", err)
		a.log.Trace().Msgf("erroneuos key ID: %s", idStr)
		return h.Res(c, fiber.StatusInternalServerError, "Internal server error")
	}

	return c.SendString(keys.Public.String())
}

// Verify verifies the received data against the private key
// Obviously this must not be exposed by an API. It's for internal use.
func (a *Service) GetPrivateKey(identifier uuid.UUID) (key *age.X25519Identity, err error) {
	// get the private key from the secrets storage
	var privKey string
	err = a.secStorage.QueryRow(
		fmt.Sprintf(`SELECT private FROM keys WHERE id='%s'`, identifier.String())).Scan(&privKey)
	if err != nil {
		a.log.Error().Msgf("error retrieving private key from secrets storage: %v", err)
		return nil, err
	}

	key, err = age.ParseX25519Identity(privKey)
	if err != nil {
		a.log.Error().Msgf("error parsing private key from secrets storage: %v", err)
		return nil, err
	}
	return key, nil
}
