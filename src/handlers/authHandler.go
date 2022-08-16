package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/src/controllers"
)

type AuthHandler struct {
	controller controllers.IAuthController
}

func NewAuthHandler(controller controllers.IAuthController) *AuthHandler {
	return &AuthHandler{controller: controller}
}

func (a *AuthHandler) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err_ := ctx.Cookie("authToken")
		if err_ != nil {
			ctx.JSON(http.StatusUnauthorized, "no token")
			return
		}

		user, err := a.controller.Auth(ctx, token)
		if err.CheckAndResponse(ctx) {
			_ = ctx.AbortWithError(err.Code, err.Error)
			return
		}

		ctx.Set("user", user)
	}
}
