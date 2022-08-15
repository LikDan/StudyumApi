package parser

import (
	"context"
	"github.com/robfig/cron"
	"studyum/src/models"
	"studyum/src/parser/apps"
)

var Apps = []models.IParserApp{&apps.KbpApp}

func UpdateGeneralSchedule(app models.IParserApp) {
	ctx := context.Background()

	var types []models.ScheduleTypeInfo
	apps.Repository.GetScheduleTypesToParse(ctx, app.GetName(), &types)

	for _, type_ := range types {
		lessons := app.GeneralScheduleUpdate(&type_)
		apps.Repository.UpdateGeneralSchedule(ctx, lessons)
	}
}

func Update(app models.IParserApp) {
	ctx := context.Background()

	var users []models.ParseJournalUser
	apps.Repository.GetUsersToParse(ctx, app.GetName(), &users)

	for _, user := range users {
		marks := app.JournalUpdate(&user)
		apps.Repository.AddMarks(ctx, marks)
		apps.Repository.UpdateParseJournalUser(ctx, &user)
	}

	var types []models.ScheduleTypeInfo
	apps.Repository.GetScheduleTypesToParse(ctx, app.GetName(), &types)

	for _, type_ := range types {
		lessons := app.ScheduleUpdate(&type_)
		apps.Repository.AddLessons(ctx, lessons)
	}
}

func InitApps() {
	ctx := context.Background()

	for _, app := range Apps {
		var lesson models.Lesson
		apps.Repository.GetLastLesson(ctx, app.GetStudyPlaceId(), &lesson)

		app.Init(lesson)

		types := app.ScheduleTypesUpdate()
		apps.Repository.InsertScheduleTypes(ctx, types)

		updateCron := cron.New()
		if err := updateCron.AddFunc(app.GetUpdateCronPattern(), func() { Update(app) }); err != nil {
			return
		}

		updateCron.Start()
	}
}
