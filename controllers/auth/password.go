package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2i RFC recommends at least time = 3, mem = 32. Time has been increased to 4
// as higher memory cost would necessitate a dedicated instance solely for hashing
const (
	timeComplexity = 4
	mem            = 32 * 1024 // 32 MiB
	keyLen         = 32
	saltLen        = 28
)

var threads = uint8(runtime.NumCPU() - 1)

func hashWithArgon(password []byte) (fmtedHash string, err error) {
	salt, err := byteGen(saltLen)
	if err != nil {
		return "", err
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
	hashedPassword := argon2.IDKey([]byte(password), salt, timeComplexity, mem, threads, keyLen)

	return bytes.Equal(encHash, hashedPassword)
}
