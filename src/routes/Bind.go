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
	api.GET("/uptime", func(ctx *gin.Context) {
		ctx.JSON(200, "hi")
	})

	Schedule(scheduleGroup)
	Journal(journalGroup)
	User(userGroup)
}
