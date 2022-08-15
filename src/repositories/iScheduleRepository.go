package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
)

type IScheduleRepository interface {
	GetSchedule(ctx context.Context, studyPlaceId int, type_ string, typeName string, schedule *models.Schedule) *models.Error
	GetScheduleType(ctx context.Context, studyPlaceId int, type_ string) []string

	AddLesson(ctx context.Context, lesson *models.Lesson, studyPlaceId int) *models.Error
	UpdateLesson(ctx context.Context, lesson *models.Lesson, studyPlaceId int) *models.Error
	DeleteLesson(ctx context.Context, id primitive.ObjectID, studyPlaceId int) *models.Error
}
