package apps

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/parser/appDTO"
	"studyum/internal/parser/entities"
	"time"
)

type App interface {
	Init(time time.Time, client *mongo.Client)

	GetName() string
	StudyPlaceId() primitive.ObjectID
	GetUpdateCronPattern() string
	LaunchCron() bool

	ScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []appDTO.LessonDTO
	GeneralScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []appDTO.GeneralLessonDTO
	ScheduleTypesUpdate() []appDTO.ScheduleTypeInfoDTO

	JournalUpdate(user entities.JournalUser) []appDTO.MarkDTO

	OnMarkAdd(ctx context.Context, mark entities.Mark, lesson entities.Lesson, student entities.User) appDTO.ParsedInfoTypeDTO
	OnMarkEdit(ctx context.Context, mark entities.Mark, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO
	OnMarkDelete(ctx context.Context, id primitive.ObjectID)

	OnLessonAdd(ctx context.Context, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO
	OnLessonEdit(ctx context.Context, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO
	OnLessonDelete(ctx context.Context, lesson entities.Lesson)

	GetSignUpDataByCode(ctx context.Context, code string) (appDTO.SignUpCode, error)

	CommitUpdate()
}
