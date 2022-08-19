package application

import (
	"context"
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

	OnMarkAdd(ctx context.Context, mark entities.Mark, lesson entities.Lesson) map[string]any
	OnMarkEdit(ctx context.Context, mark entities.Mark, lesson entities.Lesson) map[string]any
	OnMarkDelete(ctx context.Context, mark entities.Mark, lesson entities.Lesson)

	OnLessonAdd(ctx context.Context, lesson entities.Lesson) map[string]any
	OnLessonEdit(ctx context.Context, lesson entities.Lesson) map[string]any
	OnLessonDelete(ctx context.Context, lesson entities.Lesson)

	CommitUpdate()
	Init(lesson entities.Lesson)
}
