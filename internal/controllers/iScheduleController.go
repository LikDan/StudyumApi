package controllers

import (
	"context"
	"studyum/internal/entities"
)

type IScheduleController interface {
	GetSchedule(ctx context.Context, type_ string, typeName string, user entities.User) (entities.Schedule, error)
	GetUserSchedule(ctx context.Context, user entities.User) (entities.Schedule, error)

	GetScheduleTypes(ctx context.Context, user entities.User) entities.Types

	AddLesson(ctx context.Context, lesson entities.Lesson, user entities.User) error
	UpdateLesson(ctx context.Context, lesson entities.Lesson, user entities.User) error
	DeleteLesson(ctx context.Context, idHex string, user entities.User) error
}
