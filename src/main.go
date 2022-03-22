package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	h "studyium/api"
	"studyium/api/journal"
	"studyium/api/parser"
	"studyium/api/schedule"
	"studyium/api/user"
	"studyium/db"
	"studyium/firebase"
	"time"
)

type Log struct {
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
	Time  time.Time `json:"time" time_format:"2006-01-02"`
}

func indexHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "hi"})
}

func logHandler(ctx *gin.Context) {
	startTime, sErr := time.Parse("2006-01-02T15:04:05", ctx.Query("startTime"))
	endTime, eErr := time.Parse("2006-01-02T15:04:05", ctx.Query("endTime"))

	if sErr != nil || eErr != nil {
		print("error")
	}

	if !startTime.IsZero() {
		startTime = startTime.Add(time.Hour * -3)
	}
	if !endTime.IsZero() {
		endTime = endTime.Add(time.Hour * -3)
	}

	r, err := os.Open("testlogfile.jsonl")
	if err != nil {
		print(err.Error())
	}

	var allLogs []Log
	h.DecodeJsonLines(r, &allLogs)

	var logs []Log

	for _, l := range allLogs {
		if (startTime.IsZero() || l.Time.After(startTime)) && (endTime.IsZero() || l.Time.Before(endTime)) {
			logs = append(logs, l)
		}
	}
	ctx.JSON(200, logs)
}

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	db.Init()

	firebase.InitFirebaseApp()
	parser.Launch()

	r := gin.Default()

	r.GET("/api", indexHandler)

	api := r.Group("/api")
	api.GET("/log", logHandler)

	userGroup := api.Group("/user")
	journalGroup := api.Group("/journal")
	scheduleGroup := api.Group("/schedule")
	journalTeacherGroup := journalGroup.Group("/teachers")

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
