package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	apps "studyum/internal/apps/controllers"
	auth "studyum/internal/auth/entities"
	"studyum/internal/general/controllers"
	"studyum/internal/schedule/controllers/validators"
	"studyum/internal/schedule/dto"
	"studyum/internal/schedule/entities"
	"studyum/internal/schedule/repositories"
	"studyum/internal/utils"
)

type GeneralLessonsController = utils.ICRUDControllerWithUser[entities.GeneralLesson, dto.AddGeneralLessonDTO, string]

type generalLessonsController struct {
	studyPlacesController controllers.Controller
	repository            repositories.GeneralLessonsRepository
	apps                  apps.Controller
	validator             validators.Validator
}

func NewGeneralLessonController(repository repositories.GeneralLessonsRepository, studyPlacesController controllers.Controller, apps apps.Controller, validator validators.Validator) GeneralLessonsController {
	return &generalLessonsController{apps: apps, repository: repository, studyPlacesController: studyPlacesController, validator: validator}
}

func (s *generalLessonsController) GetByID(ctx context.Context, user auth.User, idHex string) (entities.GeneralLesson, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return entities.GeneralLesson{}, errors.Wrap(NotValidParams, "id")
	}

	return s.repository.GetByID(ctx, user.StudyPlaceInfo.ID, id)
}

func (s *generalLessonsController) Add(ctx context.Context, user auth.User, addDTO dto.AddGeneralLessonDTO) (entities.GeneralLesson, error) {
	if err := s.validator.AddGeneralLesson(addDTO); err != nil {
		return entities.GeneralLesson{}, err
	}

	lesson := entities.GeneralLesson{
		Id:               primitive.NewObjectID(),
		StudyPlaceId:     user.StudyPlaceInfo.ID,
		PrimaryColor:     addDTO.PrimaryColor,
		SecondaryColor:   addDTO.SecondaryColor,
		SubjectID:        addDTO.SubjectID,
		GroupID:          addDTO.GroupID,
		TeacherID:        addDTO.TeacherID,
		RoomID:           addDTO.RoomID,
		LessonIndex:      addDTO.LessonIndex,
		DayIndex:         addDTO.DayIndex,
		WeekIndex:        addDTO.WeekIndex,
		StartTimeMinutes: addDTO.StartTimeMinutes,
		EndTimeMinutes:   addDTO.StartTimeMinutes,
	}

	if err := s.repository.Add(ctx, lesson); err != nil {
		return entities.GeneralLesson{}, err
	}

	s.apps.AsyncEvent(user.StudyPlaceInfo.ID, "AddGeneralLesson", lesson)

	return s.repository.GetByID(ctx, user.StudyPlaceInfo.ID, lesson.Id)
}

func (s *generalLessonsController) Update(ctx context.Context, user auth.User, idHex string, updateDTO dto.AddGeneralLessonDTO) (entities.GeneralLesson, error) {
	if err := s.validator.UpdateGeneralLesson(updateDTO); err != nil {
		return entities.GeneralLesson{}, err
	}

	lessonID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return entities.GeneralLesson{}, err
	}

	lesson := entities.GeneralLesson{
		Id:               lessonID,
		StudyPlaceId:     user.StudyPlaceInfo.ID,
		PrimaryColor:     updateDTO.PrimaryColor,
		SecondaryColor:   updateDTO.SecondaryColor,
		SubjectID:        updateDTO.SubjectID,
		GroupID:          updateDTO.GroupID,
		TeacherID:        updateDTO.TeacherID,
		RoomID:           updateDTO.RoomID,
		LessonIndex:      updateDTO.LessonIndex,
		DayIndex:         updateDTO.DayIndex,
		WeekIndex:        updateDTO.WeekIndex,
		StartTimeMinutes: updateDTO.StartTimeMinutes,
		EndTimeMinutes:   updateDTO.EndTimeMinutes,
	}

	err = s.repository.Update(ctx, user.StudyPlaceInfo.ID, lesson)

	s.apps.AsyncEvent(user.StudyPlaceInfo.ID, "UpdateGeneralLesson", lesson)

	return s.repository.GetByID(ctx, user.StudyPlaceInfo.ID, lesson.Id)
}

func (s *generalLessonsController) DeleteByID(ctx context.Context, user auth.User, idHex string) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return errors.Wrap(NotValidParams, "id")
	}

	lesson, err := s.repository.GetByID(ctx, user.StudyPlaceInfo.ID, id)
	if err != nil {
		return err
	}

	if err = s.repository.DeleteByID(ctx, user.StudyPlaceInfo.ID, id); err != nil {
		return err
	}

	s.apps.Event(user.StudyPlaceInfo.ID, "DeleteGeneralLesson", lesson)

	return nil
}
