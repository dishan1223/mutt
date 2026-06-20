package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/dishan1223/mutt/consts"
	"github.com/gofiber/fiber/v3/log"
)

func GenerateAPIKey() (string, string) {
	// make() is used to create a new slice of bytes with a length of consts.API_KEY_BYTES.
	b := make([]byte, consts.API_KEY_BYTES)

	// We are using rand.Read() to fill the byte slice with random data.
	// This is a cryptographically secure way to generate random bytes.
	_, err := rand.Read(b)
	if err != nil {
		log.Error("Failed To Generate API Key", "error", err)
		return "", ""
	}

	raw := hex.EncodeToString(b)

	return raw, HashAPIKey(raw)
}

func HashAPIKey(Key string) string {
	h := sha256.Sum256([]byte(Key))
	return hex.EncodeToString(h[:])
}
