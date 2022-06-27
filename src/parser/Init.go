package parser

import (
	"github.com/robfig/cron"
	"studyum/src/db"
	"studyum/src/models"
	apps2 "studyum/src/parser/apps"
)

var Apps = []models.IParserApp{&apps2.KbpApp}

func UpdateGeneralSchedule(app models.IParserApp) {
	var types []models.ScheduleTypeInfo
	db.GetScheduleTypesToParse(app.GetName(), &types)

	for _, type_ := range types {
		lessons := app.GeneralScheduleUpdate(&type_)
		db.UpdateGeneralSchedule(lessons)
	}
}

func Update(app models.IParserApp) {
	var users []models.ParseJournalUser
	db.GetUsersToParse(app.GetName(), &users)

	for _, user := range users {
		marks := app.JournalUpdate(&user)
		db.AddMarks(marks)
		db.UpdateParseJournalUser(&user)
	}

	var types []models.ScheduleTypeInfo
	db.GetScheduleTypesToParse(app.GetName(), &types)

	for _, type_ := range types {
		lessons := app.ScheduleUpdate(&type_)
		db.AddLessons(lessons)
	}
}

func InitApps() {
	for _, app := range Apps {
		var lesson models.Lesson
		db.GetLastLesson(app.GetStudyPlaceId(), &lesson)

		app.Init(lesson)

		types := app.ScheduleTypesUpdate()
		db.InsertScheduleTypes(types)

		updateCron := cron.New()
		if err := updateCron.AddFunc(app.GetUpdateCronPattern(), func() { Update(app) }); err != nil {
			return
		}

		updateCron.Start()
	}
}
