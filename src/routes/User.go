package routes

import (
	"github.com/gin-gonic/gin"
	"studyum/src/controllers"
	"studyum/src/models"
)

var UserController controllers.IUserController

func User(root *gin.RouterGroup) {
	root.GET("", Auth(), UserController.GetUser)
	root.PUT("", Auth(), UserController.UpdateUser)

	root.PUT("login", UserController.LoginUser)
	root.POST("signup", UserController.SignUpUser)

	root.PUT("signup/stage1", Auth(), UserController.SignUpUserStage1)

	root.GET("auth/:oauth", UserController.OAuth2)
	root.GET("callback", UserController.CallbackOAuth2)
	root.PUT("auth/token", UserController.PutAuthToken)

	root.DELETE("signout", UserController.SignOutUser)
	root.DELETE("revoke", UserController.RevokeToken)
}

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var user models.User
		err := UserController.AuthUserViaContext(ctx, &user)
		if err.Check() {
			_ = ctx.Error(err.Error)
			return
		}

		ctx.Set("user", user)
	}
}
