package routes

import (
	"github.com/gin-gonic/gin"
	"studyum/src/api/schedule"
	"studyum/src/controllers"
)

func Schedule(root *gin.RouterGroup) {
	root.GET(":type/:name", controllers.GetSchedule)
	root.GET("my", controllers.GetMySchedule)
	root.GET("getTypes", controllers.GetScheduleTypes)

	root.POST("", controllers.AddLesson)
	root.DELETE(":id", controllers.DeleteLesson)

	root.POST("update", controllers.Update)

	schedule.BuildRequests(root)
}
