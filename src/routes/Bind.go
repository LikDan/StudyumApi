package routes

import (
	"github.com/gin-gonic/gin"
	"studyum/src/controllers"
)

func Bind(engine *gin.Engine) {
	api := engine.Group("/api")

	userGroup := api.Group("/user")
	journalGroup := api.Group("/journal")
	scheduleGroup := api.Group("/schedule")

	api.GET("/studyPlaces", controllers.GetStudyPlaces)

	Schedule(scheduleGroup)
	Journal(journalGroup)
	User(userGroup)
}
