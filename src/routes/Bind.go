package routes

import (
	"github.com/gin-gonic/gin"
	"studyum/src/api/journal"
	logApi "studyum/src/api/log"
	"studyum/src/api/parser"
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
	User(userGroup)

	logApi.BuildRequests(logGroup)
	journal.BuildRequests(journalGroup)

	api.GET("/stopPrimaryUpdates", parser.StopPrimaryCron)
	api.GET("/launchPrimaryUpdates", parser.LaunchPrimaryCron)
	api.GET("/info", parser.GetInfo)
}
