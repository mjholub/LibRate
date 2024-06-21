package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/lib/redist"
)

// Argon2i RFC recommends at least time = 3, mem = 32. Time has been increased to 4
// as higher memory cost would necessitate a dedicated instance solely for hashing
const (
	timeComplexity = 4
	mem            = 32 * 1024 // 32 MiB
	keyLen         = 32
	saltLen        = 28
)

// @Summary Change password
// @Description Change the password for the currently logged in user
// @Tags auth,accounts,updating,settings
// @Accept json
// @Produce json
// @Param old body string true "The old password"
// @Param new body string true "The new password"
// @Param X-CSRF-Token header string true "CSRF protection token"
// @Param Authorization header string true "JWT token"
// @Router /authenticate/password [patch]
func (a *Service) ChangePassword(c *fiber.Ctx) error {
	a.log.Debug().Msg("Change password request")
	memberName := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["member_name"].(string)
	if memberName == "" {
		return h.Res(c, fiber.StatusUnauthorized, "Not logged in")
	}

	passHash, err := a.ms.GetPassHash(c.Context(), "", memberName)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to retrieve password hash")
	}
	oldPass := c.Params("old")

	if !checkArgonPassword(oldPass, passHash) {
		return h.Res(c, fiber.StatusUnauthorized, "Invalid password")
	}

	newPass := c.Params("new")
	_, err = redist.CheckPasswordEntropy(newPass)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Password does not meet complexity requirements")
	}

	hash, err := hashWithArgon([]byte(newPass))
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to hash password")
	}

	err = a.ms.UpdatePassword(c.Context(), memberName, hash)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to update password")
	}
	return nil
}

func hashWithArgon(password []byte) (fmtedHash string, err error) {
	salt, err := byteGen(saltLen)
	if err != nil {
		return "", err
	}

	var threads uint8
	numCPU := runtime.NumCPU()
	if numCPU > 1 {
		threads = uint8(numCPU - 1)
	} else {
		threads = 1
	}

	hash := argon2.IDKey(password, salt, timeComplexity, mem, threads, keyLen)

	encSalt := base64.RawStdEncoding.EncodeToString(salt)
	encHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2i$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version,
		mem, timeComplexity, threads, encSalt, encHash), nil
}

func byteGen(n uint32) ([]byte, error) {
	randBytes := make([]byte, n)
	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, fmt.Errorf("error generating random bytes: %v", err)
	}
	return randBytes, nil
}

func checkArgonPassword(password, hash string) bool {
	// Split hash into components
	hashParts := strings.Split(hash, "$")
	if len(hashParts) != 6 {
		return false
	}
	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(hashParts[4])
	if err != nil {
		return false
	}
	encHash, err := base64.RawStdEncoding.DecodeString(hashParts[5])
	if err != nil {
		return false
	}
	// Hash password

	var threads uint8
	numCPU := runtime.NumCPU()
	if numCPU > 1 {
		threads = uint8(numCPU - 1)
	} else {
		threads = 1
	}
	hashedPassword := argon2.IDKey([]byte(password), salt, timeComplexity, mem, threads, keyLen)

	return bytes.Equal(encHash, hashedPassword)
}
