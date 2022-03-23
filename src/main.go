package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"studyium/api/journal"
	logApi "studyium/api/log"
	"studyium/api/parser"
	"studyium/api/schedule"
	"studyium/api/user"
	"studyium/db"
	"studyium/firebase"
	"time"
)

func indexHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "hi"})
}

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	db.Init()

	firebase.InitFirebaseApp()
	parser.Launch()

	logApi.InitLog()

	r := gin.Default()

	r.GET("/api", indexHandler)
	defer logApi.CloseLogFile()

	api := r.Group("/api")

	logGroup := api.Group("/log")
	userGroup := api.Group("/user")
	journalGroup := api.Group("/journal")
	scheduleGroup := api.Group("/schedule")
	journalTeacherGroup := journalGroup.Group("/teachers")

	logApi.BuildRequests(logGroup)

	user.BuildRequests(userGroup)
	schedule.BuildRequests(scheduleGroup, api)
	journal.BuildRequests(journalTeacherGroup)

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
