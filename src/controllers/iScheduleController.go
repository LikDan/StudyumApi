package controllers

import (
	"context"
	"studyum/src/models"
)

type IScheduleController interface {
	GetSchedule(ctx context.Context, type_ string, typeName string, user models.User) (models.Schedule, *models.Error)
	GetUserSchedule(ctx context.Context, user models.User) (models.Schedule, *models.Error)

	GetScheduleTypes(ctx context.Context, user models.User) models.Types

	AddLesson(ctx context.Context, lesson models.Lesson, user models.User) *models.Error
	UpdateLesson(ctx context.Context, lesson models.Lesson, user models.User) *models.Error
	DeleteLesson(ctx context.Context, idHex string, user models.User) *models.Error
}
