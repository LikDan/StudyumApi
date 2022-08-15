package routes

import (
	"github.com/gin-gonic/gin"
	"studyum/src/controllers"
)

var ScheduleController controllers.IScheduleController

func Schedule(root *gin.RouterGroup) {
	root.GET(":type/:name", Auth(), ScheduleController.GetSchedule)
	root.GET("", Auth(), ScheduleController.GetMySchedule)
	root.GET("getTypes", Auth(), ScheduleController.GetScheduleTypes)

	root.POST("", Auth(), ScheduleController.AddLesson)
	root.PUT("", Auth(), ScheduleController.UpdateLesson)
	root.DELETE(":id", Auth(), ScheduleController.DeleteLesson)

	root.POST("update", Auth(), ScheduleController.UpdateSchedule)
	root.POST("updateGeneral", Auth(), ScheduleController.UpdateGeneralSchedule)
}
