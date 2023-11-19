package auth

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"filippo.io/age"

	"codeberg.org/mjh/LibRate/internal/crypt"
	h "codeberg.org/mjh/LibRate/internal/handlers"

	"github.com/goccy/go-json"
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

	keysMap := map[string]string{
		private: keys.Private.String(),
		public:  keys.Public.String(),
	}

	unwrapped := keyData{
		identifier:       idStr,
		privPubNestedMap: keysMap,
	}

	keyData, err := json.MarshalIndent(unwrapped, "", "  ")
	if err != nil {
		a.log.Error().Msgf("error marshalling X25519 key data: %v", err)
		return h.Res(c, fiber.StatusInternalServerError, "Internal server error")
	}

	// we need to also append a comma to the end of the JSON object

	// save the data to secrets storage to be retrieved later
	// pubKey is the key seeded upon application start, used to encrypt the temporary secrets
	// storage
	pubKey := a.identity.Recipient()

	secFile, err := os.Open(path.Join("tmp", "sec.json"))
	if err != nil {
		a.log.Error().Msgf("error opening secrets storage: %v", err)
		return h.Res(c, fiber.StatusInternalServerError, "Internal server error")
	}
	defer func() {
		err = secFile.Close()
		if err != nil {
			a.log.Warn().Msgf("file not closed: tmp/sec.json: %v", err)
		}
	}()

	w, err := age.Encrypt(secFile, pubKey)
	if err != nil {
		a.log.Error().Msgf("error creating encrypted buffer: %v", err)
		return h.Res(c, fiber.StatusInternalServerError, "Internal server error")
	}

	if _, err := io.WriteString(w, string(keyData)); err != nil {
		a.log.Error().Msgf("error writing X25519 key details: %v", err)
		return h.Res(c, fiber.StatusInternalServerError, "Internal server error")
	}
	if err := w.Close(); err != nil {
		a.log.Error().Msgf("failed to close encrypted buffer: %v", err)
		return h.Res(c, fiber.StatusInternalServerError, "Internal server error")
	}

	return c.SendString(keys.Public.String())
}

// Verify verifies the received data against the private key
// Obviously this must not be exposed by an API. It's for internal use.
// FIXME: this will probably scale poorly, research some encrypted database solutions
func (a *Service) GetPrivateKey(identifier uuid.UUID) (key *age.X25519Identity, err error) {
	f, err := os.Open(filepath.Join("tmp", "sec.json"))
	if err != nil {
		return nil, fmt.Errorf("error opening secrets storage: %w", err)
	}
	// get an io.Reader
	r, err := age.Decrypt(f, a.identity)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secrets storage: %w", err)
	}

	secretsFileBuf := &bytes.Buffer{}
	if _, err := io.Copy(secretsFileBuf, r); err != nil {
		return nil, fmt.Errorf("failed to read secrets storage: %w", err)
	}

	// unmarshal the JSON data to get the private-public key pair
	// The data should look like this:
	// "<uuid>": {"<private key>", "<public key>"},
	// Then. we can use the private key to decrypt the data
	identifierStr := identifier.String()

	var keyDataAll []keyData
	if err := json.Unmarshal(secretsFileBuf.Bytes(), &keyDataAll); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets data: %w", err)
	}
	for i := range keyDataAll {
		if keyDataAll[i].identifier == identifierStr {
			privKey, err := age.ParseX25519Identity(keyDataAll[i].privPubNestedMap["private"])
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key: %w", err)
			}
			return privKey, nil
		}
		if i == len(keyDataAll)-1 {
			return nil, fmt.Errorf("failed to find private key for identifier %s", identifierStr)
		}
	}

	return nil, errors.New("unknown error")
}
