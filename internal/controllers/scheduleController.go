package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
	"studyum/internal/repositories"
)

type ScheduleController interface {
	GetSchedule(ctx context.Context, type_ string, typeName string, user entities.User) (entities.Schedule, error)
	GetUserSchedule(ctx context.Context, user entities.User) (entities.Schedule, error)

	GetScheduleTypes(ctx context.Context, user entities.User) entities.Types

	AddLesson(ctx context.Context, lesson entities.Lesson, user entities.User) error
	UpdateLesson(ctx context.Context, lesson entities.Lesson, user entities.User) error
	DeleteLesson(ctx context.Context, idHex string, user entities.User) error
}

type scheduleController struct {
	repository repositories.ScheduleRepository
}

func NewScheduleController(repository repositories.ScheduleRepository) ScheduleController {
	return &scheduleController{repository: repository}
}

func (s *scheduleController) GetSchedule(ctx context.Context, type_ string, typeName string, user entities.User) (entities.Schedule, error) {
	if type_ == "" || typeName == "" {
		return entities.Schedule{}, NotValidParams
	}

	return s.repository.GetSchedule(ctx, user.StudyPlaceId, type_, typeName)
}

func (s *scheduleController) GetUserSchedule(ctx context.Context, user entities.User) (entities.Schedule, error) {
	return s.repository.GetSchedule(ctx, user.StudyPlaceId, user.Type, user.TypeName)
}

func (s *scheduleController) GetScheduleTypes(ctx context.Context, user entities.User) entities.Types {
	return entities.Types{
		Groups:   s.repository.GetScheduleType(ctx, user.StudyPlaceId, "group"),
		Teachers: s.repository.GetScheduleType(ctx, user.StudyPlaceId, "teacher"),
		Subjects: s.repository.GetScheduleType(ctx, user.StudyPlaceId, "subject"),
		Rooms:    s.repository.GetScheduleType(ctx, user.StudyPlaceId, "room"),
	}
}

func (s *scheduleController) AddLesson(ctx context.Context, lesson entities.Lesson, user entities.User) error {
	lesson.StudyPlaceId = user.StudyPlaceId
	return s.repository.AddLesson(ctx, lesson)
}

func (s *scheduleController) UpdateLesson(ctx context.Context, lesson entities.Lesson, user entities.User) error {
	return s.repository.UpdateLesson(ctx, lesson, user.StudyPlaceId)
}

func (s *scheduleController) DeleteLesson(ctx context.Context, idHex string, user entities.User) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return errors.Wrap(NotValidParams, "id")
	}

	return s.repository.DeleteLesson(ctx, id, user.StudyPlaceId)
}
