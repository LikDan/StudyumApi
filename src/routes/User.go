package routes

import (
	"github.com/gin-gonic/gin"
	"studyum/src/api/user"
	"studyum/src/controllers"
	"studyum/src/controllers/oauth2"
)

func User(root *gin.RouterGroup) {
	root.GET("", controllers.GetUser)
	root.PUT("", controllers.UpdateUser)

	root.PUT("login", controllers.LoginUser)
	root.POST("signup", controllers.SignUpUser)

	root.PUT("signup/stage1", controllers.SignUpUserStage1)

	root.GET("auth/:oauth", oauth2.OAuth2)
	root.GET("callback", oauth2.CallbackOAuth2)

	root.DELETE("signout", controllers.SignOutUser)
	root.DELETE("revoke", controllers.RevokeToken)

	user.BuildRequests(root)
}
