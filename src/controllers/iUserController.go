package controllers

import (
	"github.com/gin-gonic/gin"
	"studyum/src/models"
)

type IUserController interface {
	AuthUserViaContext(ctx *gin.Context, user *models.User, permissions ...string) *models.Error

	GetUser(ctx *gin.Context)
	UpdateUser(ctx *gin.Context)

	LoginUser(ctx *gin.Context)
	SignUpUser(ctx *gin.Context)
	SignUpUserStage1(ctx *gin.Context)
	SignOutUser(ctx *gin.Context)

	RevokeToken(ctx *gin.Context)

	OAuth2(ctx *gin.Context)
	PutAuthToken(ctx *gin.Context)
	CallbackOAuth2(ctx *gin.Context)
}
