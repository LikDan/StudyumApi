package apps

import (
	"context"
	"studyum/internal/parser/dto"
	"studyum/internal/parser/entities"
	"time"
)

type App interface {
	Init(time time.Time)

	GetName() string
	StudyPlaceId() int
	GetUpdateCronPattern() string

	ScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []dto.LessonDTO
	GeneralScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []dto.GeneralLessonDTO
	ScheduleTypesUpdate() []dto.ScheduleTypeInfoDTO

	JournalUpdate(user entities.JournalUser) []dto.MarkDTO

	OnMarkAdd(ctx context.Context, mark entities.Mark, lesson entities.Lesson) entities.ParsedInfoType
	OnMarkEdit(ctx context.Context, mark entities.Mark, lesson entities.Lesson) entities.ParsedInfoType
	OnMarkDelete(ctx context.Context, mark entities.Mark, lesson entities.Lesson)

	OnLessonAdd(ctx context.Context, lesson entities.Lesson) entities.ParsedInfoType
	OnLessonEdit(ctx context.Context, lesson entities.Lesson) entities.ParsedInfoType
	OnLessonDelete(ctx context.Context, lesson entities.Lesson)

	CommitUpdate()
}
