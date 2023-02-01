package utils

import "math/rand"

var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func RandomCode(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
