package utils

import (
	"github.com/gin-gonic/gin"
	auth "studyum/internal/auth/entities"
)

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

func HasPermission(user auth.User, permission string) bool {
	for _, permission_ := range user.StudyPlaceInfo.Permissions {
		if permission_ == permission || permission_ == "admin" {
			return true
		}
	}

	return false
}
