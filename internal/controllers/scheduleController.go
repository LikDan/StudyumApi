package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/controllers/validators"
	"studyum/internal/dto"
	"studyum/internal/entities"
	parser "studyum/internal/parser/handler"
	"studyum/internal/repositories"
	"time"
)

type ScheduleController interface {
	GetSchedule(ctx context.Context, studyPlaceID string, type_ string, typeName string, user entities.User) (entities.Schedule, error)
	GetUserSchedule(ctx context.Context, user entities.User) (entities.Schedule, error)

	GetScheduleTypes(ctx context.Context, user entities.User, idHex string) entities.Types

	AddGeneralLessons(ctx context.Context, user entities.User, lessonsDTO []dto.AddGeneralLessonDTO) ([]entities.GeneralLesson, error)
	AddLessons(ctx context.Context, user entities.User, lessonsDTO []dto.AddLessonDTO) ([]entities.Lesson, error)
	AddLesson(ctx context.Context, lesson dto.AddLessonDTO, user entities.User) (entities.Lesson, error)
	UpdateLesson(ctx context.Context, lesson dto.UpdateLessonDTO, user entities.User) error
	DeleteLesson(ctx context.Context, idHex string, user entities.User) error
	RemoveLessonBetweenDates(ctx context.Context, user entities.User, date1, date2 time.Time) error

	SaveCurrentScheduleAsGeneral(ctx context.Context, user entities.User, type_ string, typeName string) error
	SaveGeneralScheduleAsCurrent(ctx context.Context, user entities.User, date time.Time) error
}

type scheduleController struct {
	parser    parser.Handler
	validator validators.Schedule

	repository        repositories.ScheduleRepository
	generalController GeneralController
}

func NewScheduleController(parser parser.Handler, validator validators.Schedule, repository repositories.ScheduleRepository, generalController GeneralController) ScheduleController {
	return &scheduleController{parser: parser, validator: validator, repository: repository, generalController: generalController}
}

func (s *scheduleController) GetSchedule(ctx context.Context, studyPlaceIDHex string, type_ string, typeName string, user entities.User) (entities.Schedule, error) {
	if type_ == "" || typeName == "" {
		return entities.Schedule{}, NotValidParams
	}

	studyPlaceID := user.StudyPlaceID
	restricted := true
	if id, err := primitive.ObjectIDFromHex(studyPlaceIDHex); err == nil && id != user.StudyPlaceID {
		studyPlaceID = id
		restricted = false
	}

	return s.repository.GetSchedule(ctx, studyPlaceID, type_, typeName, !restricted)
}

func (s *scheduleController) GetUserSchedule(ctx context.Context, user entities.User) (entities.Schedule, error) {
	return s.repository.GetSchedule(ctx, user.StudyPlaceID, user.Type, user.TypeName, false)
}

func (s *scheduleController) GetScheduleTypes(ctx context.Context, user entities.User, idHex string) entities.Types {
	studyPlaceID := user.StudyPlaceID
	if id, err := primitive.ObjectIDFromHex(idHex); err == nil && user.StudyPlaceID != id {
		if err, _ = s.repository.GetStudyPlaceByID(ctx, id, false); err != nil {
			return entities.Types{}
		}

		studyPlaceID = id
	}

	return entities.Types{
		Groups:   s.repository.GetScheduleType(ctx, studyPlaceID, "group"),
		Teachers: s.repository.GetScheduleType(ctx, studyPlaceID, "teacher"),
		Subjects: s.repository.GetScheduleType(ctx, studyPlaceID, "subject"),
		Rooms:    s.repository.GetScheduleType(ctx, studyPlaceID, "room"),
	}
}

func (s *scheduleController) AddGeneralLessons(ctx context.Context, user entities.User, lessonsDTO []dto.AddGeneralLessonDTO) ([]entities.GeneralLesson, error) {
	lessons := make([]entities.GeneralLesson, 0, len(lessonsDTO))
	for _, lessonDTO := range lessonsDTO {
		if err := s.validator.AddGeneralLesson(lessonDTO); err != nil {
			return nil, err
		}

		lesson := entities.GeneralLesson{
			Id:             primitive.NewObjectID(),
			StudyPlaceId:   user.StudyPlaceID,
			PrimaryColor:   lessonDTO.PrimaryColor,
			SecondaryColor: lessonDTO.SecondaryColor,
			StartTime:      lessonDTO.StartTime,
			EndTime:        lessonDTO.EndTime,
			Subject:        lessonDTO.Subject,
			Group:          lessonDTO.Group,
			Teacher:        lessonDTO.Teacher,
			Room:           lessonDTO.Room,
			DayIndex:       lessonDTO.DayIndex,
			WeekIndex:      lessonDTO.WeekIndex,
		}
		lessons = append(lessons, lesson)
	}

	if err := s.repository.AddGeneralLessons(ctx, lessons); err != nil {
		return nil, err
	}

	return lessons, nil
}

func (s *scheduleController) AddLessons(ctx context.Context, user entities.User, lessonsDTO []dto.AddLessonDTO) ([]entities.Lesson, error) {
	lessons := make([]entities.Lesson, 0, len(lessonsDTO))
	for _, lessonDTO := range lessonsDTO {
		if err := s.validator.AddLesson(lessonDTO); err != nil {
			return nil, err
		}

		lesson := entities.Lesson{
			Id:             primitive.NewObjectID(),
			StudyPlaceId:   user.StudyPlaceID,
			PrimaryColor:   lessonDTO.PrimaryColor,
			SecondaryColor: lessonDTO.SecondaryColor,
			Type:           lessonDTO.Type,
			StartDate:      lessonDTO.StartDate,
			EndDate:        lessonDTO.EndDate,
			Subject:        lessonDTO.Subject,
			Group:          lessonDTO.Group,
			Teacher:        lessonDTO.Teacher,
			Room:           lessonDTO.Room,
		}

		if err := s.repository.RemoveGroupLessonBetweenDates(ctx, lesson.StartDate, lesson.EndDate, user.StudyPlaceID, lesson.Group); err != nil {
			return nil, err
		}

		if lesson.Subject == "" {
			continue
		}

		lessons = append(lessons, lesson)
	}

	if err := s.repository.AddLessons(ctx, lessons); err != nil {
		return nil, err
	}

	return lessons, nil
}

func (s *scheduleController) AddLesson(ctx context.Context, dto dto.AddLessonDTO, user entities.User) (entities.Lesson, error) {
	if err := s.validator.AddLesson(dto); err != nil {
		return entities.Lesson{}, err
	}

	lesson := entities.Lesson{
		StudyPlaceId:   user.StudyPlaceID,
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
	if err := s.validator.UpdateLesson(dto); err != nil {
		return err
	}

	lesson := entities.Lesson{
		Id:             dto.Id,
		StudyPlaceId:   user.StudyPlaceID,
		PrimaryColor:   dto.PrimaryColor,
		SecondaryColor: dto.SecondaryColor,
		EndDate:        dto.EndDate,
		StartDate:      dto.StartDate,
		Subject:        dto.Subject,
		Group:          dto.Group,
		Teacher:        dto.Teacher,
		Room:           dto.Room,
		Type:           dto.Type,
		Title:          dto.Title,
		Homework:       dto.Homework,
		Description:    dto.Description,
	}

	go s.parser.EditLesson(lesson)

	err, studyPlace := s.repository.GetStudyPlaceByID(ctx, user.StudyPlaceID, false)
	if err != nil {
		return err
	}

	var lessonType entities.LessonType
	for _, lessonType_ := range studyPlace.LessonTypes {
		if lessonType_.Type == lesson.Type {
			lessonType = lessonType_
			break
		}
	}

	marks := make([]string, len(lessonType.Marks)+len(lessonType.StandaloneMarks))
	for i, markType := range lessonType.Marks {
		marks[i] = markType.Mark
	}
	for i, markType := range lessonType.StandaloneMarks {
		marks[len(lessonType.Marks)+i] = markType.Mark
	}

	if err = s.repository.FilterLessonMarks(ctx, lesson.Id, marks); err != nil {
		return err
	}

	return s.repository.UpdateLesson(ctx, lesson)
}

func (s *scheduleController) DeleteLesson(ctx context.Context, idHex string, user entities.User) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return errors.Wrap(NotValidParams, "id")
	}

	lesson, err := s.repository.FindAndDeleteLesson(ctx, id, user.StudyPlaceID)
	if err != nil {
		return err
	}

	go s.parser.DeleteLesson(lesson)

	return nil
}

func (s *scheduleController) RemoveLessonBetweenDates(ctx context.Context, user entities.User, date1, date2 time.Time) error {
	if !date1.Before(date2) {
		return errors.Wrap(validators.ValidationError, "start time is after end time")
	}

	return s.repository.RemoveLessonBetweenDates(ctx, date1, date2, user.StudyPlaceID)
}

func (s *scheduleController) SaveCurrentScheduleAsGeneral(ctx context.Context, user entities.User, type_ string, typeName string) error {
	schedule, err := s.repository.GetSchedule(ctx, user.StudyPlaceID, type_, typeName, false)
	if err != nil {
		return err
	}

	lessons := make([]entities.GeneralLesson, len(schedule.Lessons))
	for i, lesson := range schedule.Lessons {
		_, weekIndex := lesson.StartDate.ISOWeek()
		dayIndex := int(lesson.StartDate.Weekday()) - 1
		if dayIndex == -1 {
			dayIndex = 6
		}

		gLesson := entities.GeneralLesson{
			Id:             primitive.NewObjectID(),
			StudyPlaceId:   user.StudyPlaceID,
			EndTime:        lesson.EndDate.Format("15:04"),
			StartTime:      lesson.StartDate.Format("15:04"),
			PrimaryColor:   lesson.PrimaryColor,
			SecondaryColor: lesson.SecondaryColor,
			Subject:        lesson.Subject,
			Group:          lesson.Group,
			Teacher:        lesson.Teacher,
			Room:           lesson.Room,
			DayIndex:       dayIndex,
			WeekIndex:      weekIndex % schedule.Info.StudyPlace.WeeksCount,
		}

		lessons[i] = gLesson
	}

	if err = s.repository.RemoveGeneralLessonsByType(ctx, user.StudyPlaceID, type_, typeName); err != nil {
		return err
	}

	if err = s.repository.UpdateGeneralSchedule(ctx, lessons); err != nil {
		return err
	}

	return nil
}

func (s *scheduleController) SaveGeneralScheduleAsCurrent(ctx context.Context, user entities.User, date time.Time) error {
	startDayDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	err, studyPlace := s.generalController.GetStudyPlaceByID(ctx, user.StudyPlaceID, false)
	if err != nil {
		return err
	}

	_, week := date.ISOWeek()
	weekday := int(date.Weekday()) - 1
	if weekday == -1 {
		weekday = 6
	}

	generalLessons, err := s.repository.GetGeneralLessons(ctx, user.StudyPlaceID, week%studyPlace.WeeksCount, weekday)
	if err != nil {
		return err
	}

	lessons := make([]entities.Lesson, len(generalLessons))
	for i, generalLesson := range generalLessons {
		startDate, err := time.Parse("2006-01-02T15:04", date.Format("2006-01-02T")+generalLesson.StartTime)
		if err != nil {
			return err
		}

		endDate, err := time.Parse("2006-01-02T15:04", date.Format("2006-01-02T")+generalLesson.EndTime)
		if err != nil {
			return err
		}

		lesson := entities.Lesson{
			Id:             primitive.NewObjectID(),
			StudyPlaceId:   user.StudyPlaceID,
			PrimaryColor:   generalLesson.PrimaryColor,
			SecondaryColor: generalLesson.SecondaryColor,
			Type:           generalLesson.Type,
			StartDate:      startDate,
			EndDate:        endDate,
			Subject:        generalLesson.Subject,
			Group:          generalLesson.Group,
			Teacher:        generalLesson.Teacher,
			Room:           generalLesson.Room,
			Title:          "",
			Homework:       "",
			Description:    "",
		}

		lessons[i] = lesson
	}

	if err = s.repository.RemoveLessonBetweenDates(ctx, startDayDate, startDayDate.AddDate(0, 0, 1), user.StudyPlaceID); err != nil {
		return err
	}
	return s.repository.AddLessons(ctx, lessons)
}
