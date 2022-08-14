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
