package handlers

import "github.com/gin-gonic/gin"

type IUserHandler interface {
	GetUser(ctx *gin.Context)
	UpdateUser(ctx *gin.Context)

	LoginUser(ctx *gin.Context)
	SignUpUser(ctx *gin.Context)
	SignUpUserStage1(ctx *gin.Context)
	SignOutUser(ctx *gin.Context)

	OAuth2(ctx *gin.Context)
	CallbackOAuth2(ctx *gin.Context)
	PutAuthToken(ctx *gin.Context)

	RevokeToken(ctx *gin.Context)
}
