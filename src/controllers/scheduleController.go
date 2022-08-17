package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/entities"
	"studyum/src/repositories"
)

type ScheduleController struct {
	repository repositories.IScheduleRepository
}

func NewScheduleController(repository repositories.IScheduleRepository) *ScheduleController {
	return &ScheduleController{repository: repository}
}

func (s *ScheduleController) GetSchedule(ctx context.Context, type_ string, typeName string, user entities.User) (entities.Schedule, error) {
	if type_ == "" || typeName == "" {
		return entities.Schedule{}, NotValidParams
	}

	var schedule entities.Schedule
	if err := s.repository.GetSchedule(ctx, user.StudyPlaceId, type_, typeName, &schedule); err != nil {
		return entities.Schedule{}, err
	}

	return schedule, nil
}

func (s *ScheduleController) GetUserSchedule(ctx context.Context, user entities.User) (entities.Schedule, error) {
	var schedule entities.Schedule
	if err := s.repository.GetSchedule(ctx, user.StudyPlaceId, user.Type, user.TypeName, &schedule); err != nil {
		return entities.Schedule{}, err
	}

	return schedule, nil
}

func (s *ScheduleController) GetScheduleTypes(ctx context.Context, user entities.User) entities.Types {
	return entities.Types{
		Groups:   s.repository.GetScheduleType(ctx, user.StudyPlaceId, "group"),
		Teachers: s.repository.GetScheduleType(ctx, user.StudyPlaceId, "teacher"),
		Subjects: s.repository.GetScheduleType(ctx, user.StudyPlaceId, "subject"),
		Rooms:    s.repository.GetScheduleType(ctx, user.StudyPlaceId, "room"),
	}
}

func (s *ScheduleController) AddLesson(ctx context.Context, lesson entities.Lesson, user entities.User) error {
	return s.repository.AddLesson(ctx, &lesson, user.StudyPlaceId)
}

func (s *ScheduleController) UpdateLesson(ctx context.Context, lesson entities.Lesson, user entities.User) error {
	return s.repository.UpdateLesson(ctx, &lesson, user.StudyPlaceId)
}

func (s *ScheduleController) DeleteLesson(ctx context.Context, idHex string, user entities.User) error {
	if !primitive.IsValidObjectID(idHex) {
		return errors.Wrap(NotValidParams, "id")
	}

	id, _ := primitive.ObjectIDFromHex(idHex)
	if err := s.repository.DeleteLesson(ctx, id, user.StudyPlaceId); err != nil {
		return err
	}

	return nil
}
