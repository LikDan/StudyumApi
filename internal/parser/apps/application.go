package apps

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/appDTO"
	"studyum/internal/parser/entities"
	"time"
)

type App interface {
	Init(time time.Time)

	GetName() string
	StudyPlaceId() primitive.ObjectID
	GetUpdateCronPattern() string

	ScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []appDTO.LessonDTO
	GeneralScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []appDTO.GeneralLessonDTO
	ScheduleTypesUpdate() []appDTO.ScheduleTypeInfoDTO

	JournalUpdate(user entities.JournalUser) []appDTO.MarkDTO

	OnMarkAdd(ctx context.Context, mark entities.Mark, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO
	OnMarkEdit(ctx context.Context, mark entities.Mark, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO
	OnMarkDelete(ctx context.Context, mark entities.Mark, lesson entities.Lesson)

	OnLessonAdd(ctx context.Context, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO
	OnLessonEdit(ctx context.Context, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO
	OnLessonDelete(ctx context.Context, lesson entities.Lesson)

	CommitUpdate()
}
