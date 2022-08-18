package utils

import (
	"github.com/gin-gonic/gin"
	"studyum/internal/entities"
)

func GetUserViaCtx(ctx *gin.Context) entities.User {
	return GetViaCtx[entities.User](ctx, "user")
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