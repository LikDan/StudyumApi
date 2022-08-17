package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
)

type IScheduleRepository interface {
	GetSchedule(ctx context.Context, studyPlaceId int, type_ string, typeName string, schedule *models.Schedule) error
	GetScheduleType(ctx context.Context, studyPlaceId int, type_ string) []string

	AddLesson(ctx context.Context, lesson *models.Lesson, studyPlaceId int) error
	UpdateLesson(ctx context.Context, lesson *models.Lesson, studyPlaceId int) error
	DeleteLesson(ctx context.Context, id primitive.ObjectID, studyPlaceId int) error
}
