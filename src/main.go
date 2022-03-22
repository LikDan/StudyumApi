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

	r := gin.Default()

	r.GET("/api", indexHandler)

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

	err := r.Run()
	if err != nil {
		log.Fatalf("Error launching server %s", err.Error())
	}
}

/*func main() {
	log.SetFormatter(&log.JSONFormatter{})

	f, err := os.OpenFile("testlogfile.jsonl", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Info("succeeded")
	log.Warn("not correct")
	log.Error("something error")

	// A common pattern is to re-use fields between logging statements by re-using
	// the log.Entry returned from WithFields()
	contextLogger := log.WithFields(log.Fields{
		"common": "this is a common field",
		"other":  "I also should be logged always",
	})

	contextLogger.Info("I'll be logged with common and other field")
	contextLogger.Info("Me too")

	log.Trace("Something very low level.")
	log.Debug("Useful debugging information.")
	log.Info("Something noteworthy happened!")
	log.Warn("You should probably take a look at this.")
	log.Error("Something failed but I'm not quitting.")
	// Calls os.Exit(1) after logging
	log.Fatal("Bye.")
	// Calls panic() after logging
	log.Panic("I'm bailing.")
}
*/
