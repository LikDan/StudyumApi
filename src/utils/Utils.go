package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"studyum/src/models"
)

func CheckEmpty(strings ...string) bool {
	for _, s := range strings {
		if s == "" {
			return true
		}
	}
	return false
}
func CheckEmptyAndResponse(ctx *gin.Context, err *models.Error, strings ...string) bool {
	if !CheckEmpty(strings...) {
		return false
	}

	ctx.JSON(err.Code, err.Error)
	return true
}

func CheckNotEmpty(strings ...string) bool {
	for _, s := range strings {
		if s == "" {
			return false
		}
	}
	return true
}

func GenerateSecureToken() string {
	b := make([]byte, 128)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func GetUserViaCtx(ctx *gin.Context) models.User {
	return GetViaCtx[models.User](ctx, "user")
}

func GetViaCtx[G any](ctx *gin.Context, name string) G {
	var def G

	i, ok := ctx.Get(name)
	if !ok {
		return def
	}

	g, ok := i.(G)
	if !ok {
		return def
	}

	return g
}
