package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
	"studyum/src/repositories"
	"studyum/src/utils"
)

type ScheduleController struct {
	repository repositories.IScheduleRepository
}

func NewScheduleController(repository repositories.IScheduleRepository) *ScheduleController {
	return &ScheduleController{repository: repository}
}

func (s *ScheduleController) GetSchedule(ctx context.Context, type_ string, typeName string, user models.User) (models.Schedule, error) {
	if !utils.CheckNotEmpty(type_, typeName) {
		return models.Schedule{}, NotValidParams
	}

	var schedule models.Schedule
	if err := s.repository.GetSchedule(ctx, user.StudyPlaceId, type_, typeName, &schedule); err != nil {
		return models.Schedule{}, err
	}

	return schedule, nil
}

func (s *ScheduleController) GetUserSchedule(ctx context.Context, user models.User) (models.Schedule, error) {
	var schedule models.Schedule
	if err := s.repository.GetSchedule(ctx, user.StudyPlaceId, user.Type, user.TypeName, &schedule); err != nil {
		return models.Schedule{}, err
	}

	return schedule, nil
}

func (s *ScheduleController) GetScheduleTypes(ctx context.Context, user models.User) models.Types {
	return models.Types{
		Groups:   s.repository.GetScheduleType(ctx, user.StudyPlaceId, "group"),
		Teachers: s.repository.GetScheduleType(ctx, user.StudyPlaceId, "teacher"),
		Subjects: s.repository.GetScheduleType(ctx, user.StudyPlaceId, "subject"),
		Rooms:    s.repository.GetScheduleType(ctx, user.StudyPlaceId, "room"),
	}
}

func (s *ScheduleController) AddLesson(ctx context.Context, lesson models.Lesson, user models.User) error {
	return s.repository.AddLesson(ctx, &lesson, user.StudyPlaceId)
}

func (s *ScheduleController) UpdateLesson(ctx context.Context, lesson models.Lesson, user models.User) error {
	return s.repository.UpdateLesson(ctx, &lesson, user.StudyPlaceId)
}

func (s *ScheduleController) DeleteLesson(ctx context.Context, idHex string, user models.User) error {
	if !primitive.IsValidObjectID(idHex) {
		return errors.Wrap(NotValidParams, "id")
	}

	id, _ := primitive.ObjectIDFromHex(idHex)
	if err := s.repository.DeleteLesson(ctx, id, user.StudyPlaceId); err != nil {
		return err
	}

	return nil
}
