package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"studyum/src/api/journal"
	logApi "studyum/src/api/log"
	"studyum/src/api/parser"
	"studyum/src/api/schedule"
	"studyum/src/api/user"
	"studyum/src/db"
	"studyum/src/firebase"
	"time"
)

func uptimeHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "hi"})
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

	api := r.Group("/api")

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
