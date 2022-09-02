package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"studyum/internal/dto"
	"studyum/internal/entities"
	parser "studyum/internal/parser/handler"
	"studyum/internal/repositories"
)

type ScheduleController interface {
	GetPreviewSchedule(ctx context.Context, studyPlaceID string, type_ string, typeName string) (entities.Schedule, error)
	GetSchedule(ctx context.Context, type_ string, typeName string, user entities.User) (entities.Schedule, error)
	GetUserSchedule(ctx context.Context, user entities.User) (entities.Schedule, error)

	GetScheduleTypes(ctx context.Context, user entities.User) entities.Types

	AddLesson(ctx context.Context, lesson dto.AddLessonDTO, user entities.User) (entities.Lesson, error)
	UpdateLesson(ctx context.Context, lesson dto.UpdateLessonDTO, user entities.User) error
	DeleteLesson(ctx context.Context, idHex string, user entities.User) error
	SaveCurrentScheduleAsGeneral(ctx context.Context, user entities.User, type_ string, typeName string) error
}

type scheduleController struct {
	parser parser.Handler

	repository repositories.ScheduleRepository
}

func NewScheduleController(parser parser.Handler, repository repositories.ScheduleRepository) ScheduleController {
	return &scheduleController{parser: parser, repository: repository}
}

func (s *scheduleController) GetPreviewSchedule(ctx context.Context, studyPlaceID string, type_ string, typeName string) (entities.Schedule, error) {
	if studyPlaceID == "" || type_ == "" || typeName == "" {
		return entities.Schedule{}, NotValidParams
	}

	id, err := strconv.Atoi(studyPlaceID)
	if err != nil {
		return entities.Schedule{}, err
	}

	return s.repository.GetSchedule(ctx, id, type_, typeName, true)
}

func (s *scheduleController) GetSchedule(ctx context.Context, type_ string, typeName string, user entities.User) (entities.Schedule, error) {
	if type_ == "" || typeName == "" {
		return entities.Schedule{}, NotValidParams
	}

	return s.repository.GetSchedule(ctx, user.StudyPlaceId, type_, typeName, false)
}

func (s *scheduleController) GetUserSchedule(ctx context.Context, user entities.User) (entities.Schedule, error) {
	return s.repository.GetSchedule(ctx, user.StudyPlaceId, user.Type, user.TypeName, false)
}

func (s *scheduleController) GetScheduleTypes(ctx context.Context, user entities.User) entities.Types {
	return entities.Types{
		Groups:   s.repository.GetScheduleType(ctx, user.StudyPlaceId, "group"),
		Teachers: s.repository.GetScheduleType(ctx, user.StudyPlaceId, "teacher"),
		Subjects: s.repository.GetScheduleType(ctx, user.StudyPlaceId, "subject"),
		Rooms:    s.repository.GetScheduleType(ctx, user.StudyPlaceId, "room"),
	}
}

func (s *scheduleController) AddLesson(ctx context.Context, dto dto.AddLessonDTO, user entities.User) (entities.Lesson, error) {
	lesson := entities.Lesson{
		StudyPlaceId:   user.StudyPlaceId,
		PrimaryColor:   dto.PrimaryColor,
		SecondaryColor: dto.SecondaryColor,
		EndDate:        dto.EndDate,
		StartDate:      dto.StartDate,
		Subject:        dto.Subject,
		Group:          dto.Group,
		Teacher:        dto.Teacher,
		Room:           dto.Room,
	}

	id, err := s.repository.AddLesson(ctx, lesson)
	if err != nil {
		return entities.Lesson{}, err
	}

	lesson.Id = id
	go s.parser.AddLesson(lesson)

	return lesson, err
}

func (s *scheduleController) UpdateLesson(ctx context.Context, dto dto.UpdateLessonDTO, user entities.User) error {
	lesson := entities.Lesson{
		Id:           dto.Id,
		StudyPlaceId: user.StudyPlaceId,
		Subject:      dto.Subject,
		Group:        dto.Group,
		Teacher:      dto.Teacher,
		Room:         dto.Room,
		Title:        dto.Title,
		Homework:     dto.Homework,
		Description:  dto.Description,
	}

	go s.parser.EditLesson(lesson)

	return s.repository.UpdateLesson(ctx, lesson, user.StudyPlaceId)
}

func (s *scheduleController) DeleteLesson(ctx context.Context, idHex string, user entities.User) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return errors.Wrap(NotValidParams, "id")
	}

	lesson, err := s.repository.FindAndDeleteLesson(ctx, id, user.StudyPlaceId)
	if err != nil {
		return err
	}

	go s.parser.DeleteLesson(lesson)

	return nil
}

func (s *scheduleController) SaveCurrentScheduleAsGeneral(ctx context.Context, user entities.User, type_ string, typeName string) error {
	schedule, err := s.repository.GetSchedule(ctx, user.StudyPlaceId, type_, typeName, false)
	if err != nil {
		return err
	}

	lessons := make([]entities.GeneralLesson, len(schedule.Lessons))
	for i, lesson := range schedule.Lessons {
		_, weekIndex := lesson.StartDate.ISOWeek()

		gLesson := entities.GeneralLesson{
			Id:           primitive.NewObjectID(),
			StudyPlaceId: user.StudyPlaceId,
			EndTime:      lesson.EndDate.Format("15:04"),
			StartTime:    lesson.StartDate.Format("15:04"),
			Subject:      lesson.Subject,
			Group:        lesson.Group,
			Teacher:      lesson.Teacher,
			Room:         lesson.Room,
			DayIndex:     lesson.StartDate.Day(),
			WeekIndex:    weekIndex,
		}

		lessons[i] = gLesson
	}

	if err = s.repository.UpdateGeneralSchedule(ctx, lessons, type_, typeName); err != nil {
		return err
	}

	return nil
}
