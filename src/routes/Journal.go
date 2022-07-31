package routes

import (
	"github.com/gin-gonic/gin"
	"studyum/src/controllers"
)

func Journal(root *gin.RouterGroup) {
	root.GET("/options", controllers.GetJournalAvailableOptions)
	root.GET("/:group/:subject/:teacher", controllers.GetJournal)
	root.GET("", controllers.GetUserJournal)

	mark := root.Group("/mark")
	{
		mark.POST("", controllers.AddMark)
		mark.GET("", controllers.GetMark)
		mark.PUT("", controllers.UpdateMark)
		mark.DELETE("", controllers.DeleteMark)
	}
}
