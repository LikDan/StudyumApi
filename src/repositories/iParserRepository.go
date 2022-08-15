package repositories

import (
	"context"
	"studyum/src/models"
	"time"
)

type IParserRepository interface {
	GetUsersToParse(ctx context.Context, parserAppName string, users *[]models.ParseJournalUser) *models.Error
	UpdateParseJournalUser(ctx context.Context, user *models.ParseJournalUser) *models.Error

	InsertScheduleTypes(ctx context.Context, types []*models.ScheduleTypeInfo) *models.Error
	GetScheduleTypesToParse(ctx context.Context, parserAppName string, types *[]models.ScheduleTypeInfo) *models.Error

	UpdateGeneralSchedule(ctx context.Context, lessons []*models.GeneralLesson) *models.Error
	GetLessonByDate(ctx context.Context, date time.Time, name string, group string, lesson *models.Lesson)
	GetLastLesson(ctx context.Context, studyPlaceId int, lesson *models.Lesson) *models.Error
	AddLessons(ctx context.Context, lessons []*models.Lesson) *models.Error

	AddMarks(ctx context.Context, marks []*models.Mark) *models.Error
}
