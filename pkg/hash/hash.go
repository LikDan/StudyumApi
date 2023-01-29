package hash

import (
	"golang.org/x/crypto/bcrypt"
)

func Hash(s string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	return string(bytes), err
}

func CompareHashAndPassword(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
