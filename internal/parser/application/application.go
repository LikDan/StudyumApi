package application

import (
	"studyum/internal/parser/dto"
	"studyum/internal/parser/entities"
)

type App interface {
	GetName() string
	StudyPlaceId() int
	GetUpdateCronPattern() string

	ScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []dto.LessonDTO
	GeneralScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []dto.GeneralLessonDTO
	ScheduleTypesUpdate() []entities.ScheduleTypeInfo

	JournalUpdate(user entities.JournalUser) []dto.MarkDTO

	CommitUpdate()
	Init(lesson entities.Lesson)
}