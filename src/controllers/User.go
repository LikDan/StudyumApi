package controllers

import (
	"github.com/gin-gonic/gin"
	utils "studyum/src/api"
	"studyum/src/db"
	"studyum/src/models"
)

func AuthUserViaToken(token string, user *models.User, permissions ...string) *models.Error {
	var user_ models.User
	if err := db.GetUserViaToken(token, &user_); err.Error != nil {
		return err
	}

	for _, permission := range permissions {
		if !utils.SliceContains(user.Permissions, permission) {
			return models.BindErrorStr("no permission(s)", 403, utils.UNDEFINED)
		}
	}

	*user = user_
	return models.EmptyError()
}

func AuthUserViaContext(ctx *gin.Context, user *models.User, permissions ...string) *models.Error {
	token, err := ctx.Cookie("authToken")
	if err != nil {
		return models.BindError(err, 401, utils.UNDEFINED)
	}

	AuthUserViaToken(token, user, permissions...)

	return models.EmptyError()
}
