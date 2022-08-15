package routes

import (
	"github.com/gin-gonic/gin"
	"studyum/src/controllers"
)

var JournalController controllers.IJournalController

func Journal(root *gin.RouterGroup) {
	root.GET("/options", Auth(), JournalController.GetJournalAvailableOptions)
	root.GET("/:group/:subject/:teacher", Auth(), JournalController.GetJournal)
	root.GET("", Auth(), JournalController.GetUserJournal)

	mark := root.Group("/mark", Auth())
	{
		mark.POST("", JournalController.AddMark)
		mark.GET("", JournalController.GetMark)
		mark.PUT("", JournalController.UpdateMark)
		mark.DELETE("", JournalController.DeleteMark)
	}
}
