package repository

import (
	"context"
	"studyum/internal/parser/entities"
	"time"
)

type IParserRepository interface {
	GetUsersToParse(ctx context.Context, parserAppName string, users *[]entities.JournalUser) error
	UpdateParseJournalUser(ctx context.Context, user *entities.JournalUser) error

	InsertScheduleTypes(ctx context.Context, types []*entities.ScheduleTypeInfo) error
	GetScheduleTypesToParse(ctx context.Context, parserAppName string, types *[]entities.ScheduleTypeInfo) error

	UpdateGeneralSchedule(ctx context.Context, lessons []*entities.GeneralLesson) error
	GetLessonByDate(ctx context.Context, date time.Time, name string, group string) (entities.Lesson, error)
	GetLastLesson(ctx context.Context, studyPlaceId int, lesson *entities.Lesson) error
	AddLessons(ctx context.Context, lessons []*entities.Lesson) error

	AddMarks(ctx context.Context, marks []*entities.Mark) error
}
