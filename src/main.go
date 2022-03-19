package main

import (
	"github.com/gin-gonic/gin"
	"studyium/api/journal"
	"studyium/api/parser"
	"studyium/api/schedule"
	"studyium/api/user"
	"studyium/db"
	"studyium/firebase"
	"time"
)

func indexHandler(ctx *gin.Context) {
	message(ctx, "message", "hi", 200)
}

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	db.Init()

	firebase.InitFirebaseApp()
	parser.Launch()

	r := gin.Default()

	r.GET("/api", indexHandler)

	api := r.Group("/api")

	userGroup := api.Group("/user")
	journalGroup := api.Group("/journal")
	scheduleGroup := api.Group("/schedule")

	user.BuildRequests(userGroup)
	schedule.BuildRequests(scheduleGroup, api)

	api.GET("/stopPrimaryUpdates", parser.StopPrimaryCron)
	api.GET("/launchPrimaryUpdates", parser.LaunchPrimaryCron)

	api.GET("/info", parser.GetInfo)

	journalTeacherGroup := journalGroup.Group("/teachers")
	journalTeacherGroup.GET("/types", journal.GetTeacherJournalTypes)
	journalTeacherGroup.GET("/dates", journal.GetTeacherJournalSubjects)
	journalTeacherGroup.GET("/groupMembers", journal.GetGroupMembers)

	journalTeacherGroup.GET("/addMark", journal.AddMark)
	journalTeacherGroup.GET("/getMark", journal.GetMark)
	journalTeacherGroup.GET("/editMark", journal.EditMark)
	journalTeacherGroup.GET("/removeMark", journal.RemoveMark)

	journalTeacherGroup.GET("/editInfo", journal.EditInfo)

	err := r.Run()
	CheckError(err)
}
