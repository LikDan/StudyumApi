package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"studyium/src/api/journal"
	logApi "studyium/src/api/log"
	"studyium/src/api/parser"
	"studyium/src/api/schedule"
	"studyium/src/api/user"
	"studyium/src/db"
	"studyium/src/firebase"
	"time"
)

func uptimeHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "hi"})
}

func BeforeAfterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("AllowCredentials", "true")
		c.Next()
	}
}

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	db.Init()

	firebase.InitFirebaseApp()
	parser.Launch()

	logApi.InitLog()

	r := gin.Default()

	r.HEAD("/api", uptimeHandler)
	defer logApi.CloseLogFile()

	api := r.Group("/api", BeforeAfterMiddleware())

	logGroup := api.Group("/log")
	userGroup := api.Group("/user")
	journalGroup := api.Group("/journal")
	scheduleGroup := api.Group("/schedule")

	logApi.BuildRequests(logGroup)

	user.BuildRequests(userGroup)
	schedule.BuildRequests(scheduleGroup, api)
	journal.BuildRequests(journalGroup)

	api.GET("/stopPrimaryUpdates", parser.StopPrimaryCron)
	api.GET("/launchPrimaryUpdates", parser.LaunchPrimaryCron)
	api.GET("/info", parser.GetInfo)
	scheduleGroup.GET("/update", parser.UpdateSchedule)

	log.Info("Application launched")

	err := r.Run()
	if err != nil {
		log.Fatalf("Error launching server %s", err.Error())
	}
}
