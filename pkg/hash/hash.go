package hash

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
)

func Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func GenerateSecureToken() string {
	b := make([]byte, 128)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
