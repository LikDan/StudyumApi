package routes

import (
	"github.com/gin-gonic/gin"
	logApi "studyum/src/api/log"
	"studyum/src/controllers"
)

func Bind(engine *gin.Engine) {
	api := engine.Group("/api")

	logGroup := api.Group("/log")
	userGroup := api.Group("/user")
	journalGroup := api.Group("/journal")
	scheduleGroup := api.Group("/schedule")

	api.GET("/studyPlaces", controllers.GetStudyPlaces)

	Schedule(scheduleGroup)
	Journal(journalGroup)
	User(userGroup)

	logApi.BuildRequests(logGroup)
}
