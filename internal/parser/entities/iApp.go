package entities

import (
	"context"
	"time"
)

type IApp interface {
	Init(lesson Lesson)
	CommitUpdate()

	ScheduleUpdate(type_ ScheduleTypeInfo) []Lesson
	GeneralScheduleUpdate(type_ ScheduleTypeInfo) []GeneralLesson
	ScheduleTypesUpdate() []ScheduleTypeInfo
	JournalUpdate(user JournalUser, getLessonByDate func(context.Context, time.Time, string, string) (Lesson, error)) []Mark

	GetName() string
	StudyPlaceId() int
	GetUpdateCronPattern() string
}
