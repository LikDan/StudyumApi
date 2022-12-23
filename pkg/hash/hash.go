package hash

import (
	"crypto/rand"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
)

func Hash(s string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	return string(bytes), err
}

func CompareHashAndPassword(s string, s2 string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(s), []byte(s2))
	return err == nil
}

func GenerateSecureToken() string {
	b := make([]byte, 128)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
