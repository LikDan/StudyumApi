package routes

import (
	"github.com/gin-gonic/gin"
	"studyum/src/controllers"
)

func Schedule(root *gin.RouterGroup) {
	root.GET(":type/:name", controllers.GetSchedule)
	root.GET("", controllers.GetMySchedule)
	root.GET("getTypes", controllers.GetScheduleTypes)

	root.POST("", controllers.AddLesson)
	root.PUT("", controllers.UpdateLesson)
	root.DELETE(":id", controllers.DeleteLesson)

	root.POST("update", controllers.UpdateSchedule)
	root.POST("updateGeneral", controllers.UpdateGeneralSchedule)
}
