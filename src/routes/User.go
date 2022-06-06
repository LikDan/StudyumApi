package routes

import (
	"github.com/gin-gonic/gin"
	"studyum/src/api/user"
	"studyum/src/controllers"
)

func User(root *gin.RouterGroup) {
	root.PUT("login", controllers.LoginUser)
	root.POST("signup", controllers.SignUpUser)

	root.PUT("signup/stage1", controllers.SignUpUserStage1)

	root.DELETE("signout", controllers.SignOutUser)
	root.DELETE("revoke", controllers.RevokeToken)

	user.BuildRequests(root)
}
