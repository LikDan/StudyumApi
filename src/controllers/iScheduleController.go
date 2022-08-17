package controllers

import (
	"context"
	"studyum/src/models"
)

type IScheduleController interface {
	GetSchedule(ctx context.Context, type_ string, typeName string, user models.User) (models.Schedule, error)
	GetUserSchedule(ctx context.Context, user models.User) (models.Schedule, error)

	GetScheduleTypes(ctx context.Context, user models.User) models.Types

	AddLesson(ctx context.Context, lesson models.Lesson, user models.User) error
	UpdateLesson(ctx context.Context, lesson models.Lesson, user models.User) error
	DeleteLesson(ctx context.Context, idHex string, user models.User) error
}
