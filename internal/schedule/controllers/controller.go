package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	apps "studyum/internal/apps/controllers"
	auth "studyum/internal/auth/entities"
	"studyum/internal/general/controllers"
	general "studyum/internal/general/entities"
	journalEntities "studyum/internal/journal/entities"
	"studyum/internal/schedule/controllers/validators"
	dto2 "studyum/internal/schedule/dto"
	"studyum/internal/schedule/entities"
	"studyum/internal/schedule/repositories"
	"studyum/pkg/datetime"
	"time"
)

var NotValidParams = errors.New("not valid params")

type Controller interface {
	GetSchedule(ctx context.Context, user auth.User, studyPlaceID string, role string, roleName string, startDate, endDate time.Time) (entities.Schedule, error)
	GetUserSchedule(ctx context.Context, user auth.User, startDate, endDate time.Time) (entities.Schedule, error)

	GetGeneralSchedule(ctx context.Context, user auth.User, studyPlaceID string, role string, roleName string, startDate, endDate time.Time) (entities.Schedule, error)
	GetGeneralUserSchedule(ctx context.Context, user auth.User, startDate, endDate time.Time) (entities.Schedule, error)

	GetScheduleTypes(ctx context.Context, user auth.User, idHex string) entities.Types

	AddGeneralLessons(ctx context.Context, user auth.User, lessonsDTO []dto2.AddGeneralLessonDTO) ([]entities.GeneralLesson, error)
	AddLessons(ctx context.Context, user auth.User, lessonsDTO []dto2.AddLessonDTO) ([]entities.Lesson, error)

	AddLesson(ctx context.Context, lesson dto2.AddLessonDTO, user auth.User) (entities.Lesson, error)
	GetLessonByID(ctx context.Context, user auth.User, idHex string) (entities.Lesson, error)
	UpdateLesson(ctx context.Context, lesson dto2.UpdateLessonDTO, user auth.User) error
	DeleteLesson(ctx context.Context, idHex string, user auth.User) error

	GetLessonsByDateAndID(ctx context.Context, user auth.User, idHex string) ([]entities.Lesson, error)

	RemoveLessonBetweenDates(ctx context.Context, user auth.User, date1, date2 time.Time) error

	SaveCurrentScheduleAsGeneral(ctx context.Context, user auth.User, role string, roleName string) error
	SaveGeneralScheduleAsCurrent(ctx context.Context, user auth.User, date time.Time) error
}

type controller struct {
	repository repositories.Repository

	generalController controllers.Controller

	apps      apps.Controller
	validator validators.Validator
}

func NewScheduleController(repository repositories.Repository, generalController controllers.Controller, apps apps.Controller, validator validators.Validator) Controller {
	return &controller{apps: apps, validator: validator, repository: repository, generalController: generalController}
}

func (s *controller) scheduleDated(start, end time.Time) (time.Time, time.Time) {
	emptyTime := time.Time{}
	if start == emptyTime {
		start = datetime.Date().AddDate(0, 0, 1-int(time.Now().Weekday()))
	}
	if end == emptyTime {
		end = datetime.Date().AddDate(0, 0, 8-int(time.Now().Weekday()))
	}

	return start, end
}

func (s *controller) GetSchedule(ctx context.Context, user auth.User, studyPlaceIDHex string, role string, roleName string, startDate, endDate time.Time) (entities.Schedule, error) {
	if role == "" || roleName == "" {
		return entities.Schedule{}, NotValidParams
	}

	studyPlaceID := user.StudyPlaceInfo.ID
	restricted := true
	if id, err := primitive.ObjectIDFromHex(studyPlaceIDHex); err == nil && id != user.StudyPlaceInfo.ID {
		studyPlaceID = id
		restricted = false
	}

	startDate, endDate = s.scheduleDated(startDate, endDate)
	return s.repository.GetSchedule(ctx, studyPlaceID, role, roleName, startDate, endDate, false, !restricted)
}

func (s *controller) GetUserSchedule(ctx context.Context, user auth.User, startDate, endDate time.Time) (entities.Schedule, error) {
	if user.StudyPlaceInfo.Role == "" || user.StudyPlaceInfo.RoleName == "" {
		return entities.Schedule{}, NotValidParams
	}

	startDate, endDate = s.scheduleDated(startDate, endDate)
	return s.repository.GetSchedule(ctx, user.StudyPlaceInfo.ID, user.StudyPlaceInfo.Role, user.StudyPlaceInfo.RoleName, startDate, endDate, false, false)
}

func (s *controller) GetGeneralSchedule(ctx context.Context, user auth.User, studyPlaceIDHex string, role string, roleName string, startDate, endDate time.Time) (entities.Schedule, error) {
	if role == "" || roleName == "" {
		return entities.Schedule{}, NotValidParams
	}

	studyPlaceID := user.StudyPlaceInfo.ID
	restricted := true
	if id, err := primitive.ObjectIDFromHex(studyPlaceIDHex); err == nil && id != user.StudyPlaceInfo.ID {
		studyPlaceID = id
		restricted = false
	}

	startDate, endDate = s.scheduleDated(startDate, endDate)
	return s.repository.GetSchedule(ctx, studyPlaceID, role, roleName, startDate, endDate, true, !restricted)
}

func (s *controller) GetGeneralUserSchedule(ctx context.Context, user auth.User, startDate, endDate time.Time) (entities.Schedule, error) {
	startDate, endDate = s.scheduleDated(startDate, endDate)
	return s.repository.GetSchedule(ctx, user.StudyPlaceInfo.ID, user.StudyPlaceInfo.Role, user.StudyPlaceInfo.RoleName, startDate, endDate, true, false)
}

func (s *controller) GetScheduleTypes(ctx context.Context, user auth.User, idHex string) entities.Types {
	studyPlaceID := user.StudyPlaceInfo.ID
	if id, err := primitive.ObjectIDFromHex(idHex); err == nil && user.StudyPlaceInfo.ID != id {
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

func (s *controller) AddGeneralLessons(ctx context.Context, user auth.User, lessonsDTO []dto2.AddGeneralLessonDTO) ([]entities.GeneralLesson, error) {
	lessons := make([]entities.GeneralLesson, 0, len(lessonsDTO))
	for _, lessonDTO := range lessonsDTO {
		if err := s.validator.AddGeneralLesson(lessonDTO); err != nil {
			return nil, err
		}

		lesson := entities.GeneralLesson{
			Id:             primitive.NewObjectID(),
			StudyPlaceId:   user.StudyPlaceInfo.ID,
			PrimaryColor:   lessonDTO.PrimaryColor,
			SecondaryColor: lessonDTO.SecondaryColor,
			StartTime:      lessonDTO.StartTime,
			EndTime:        lessonDTO.EndTime,
			Subject:        lessonDTO.Subject,
			Group:          lessonDTO.Group,
			Teacher:        lessonDTO.Teacher,
			Room:           lessonDTO.Room,
			LessonIndex:    lessonDTO.LessonIndex,
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

func (s *controller) AddLessons(ctx context.Context, user auth.User, lessonsDTO []dto2.AddLessonDTO) ([]entities.Lesson, error) {
	lessons := make([]entities.Lesson, 0, len(lessonsDTO))
	for _, lessonDTO := range lessonsDTO {
		if err := s.validator.AddLesson(lessonDTO); err != nil {
			return nil, err
		}

		lesson := entities.Lesson{
			Id:             primitive.NewObjectID(),
			StudyPlaceId:   user.StudyPlaceInfo.ID,
			PrimaryColor:   lessonDTO.PrimaryColor,
			SecondaryColor: lessonDTO.SecondaryColor,
			Type:           lessonDTO.Type,
			LessonIndex:    lessonDTO.LessonIndex,
			StartDate:      lessonDTO.StartDate,
			EndDate:        lessonDTO.EndDate,
			Subject:        lessonDTO.Subject,
			Group:          lessonDTO.Group,
			Teacher:        lessonDTO.Teacher,
			Room:           lessonDTO.Room,
		}

		if err := s.repository.RemoveGroupLessonBetweenDates(ctx, lesson.StartDate, lesson.EndDate, user.StudyPlaceInfo.ID, lesson.Group); err != nil {
			return nil, err
		}

		if lesson.Subject == "" {
			continue
		}

		s.apps.AsyncEvent(user.StudyPlaceInfo.ID, "AddLesson", lesson)
		lessons = append(lessons, lesson)
	}

	if err := s.repository.AddLessons(ctx, lessons); err != nil {
		return nil, err
	}

	return lessons, nil
}

func (s *controller) AddLesson(ctx context.Context, addDTO dto2.AddLessonDTO, user auth.User) (entities.Lesson, error) {
	if err := s.validator.AddLesson(addDTO); err != nil {
		return entities.Lesson{}, err
	}

	lesson := entities.Lesson{
		Id:             primitive.NewObjectID(),
		StudyPlaceId:   user.StudyPlaceInfo.ID,
		PrimaryColor:   addDTO.PrimaryColor,
		SecondaryColor: addDTO.SecondaryColor,
		Type:           addDTO.Type,
		EndDate:        addDTO.EndDate,
		StartDate:      addDTO.StartDate,
		LessonIndex:    addDTO.LessonIndex,
		Subject:        addDTO.Subject,
		Group:          addDTO.Group,
		Teacher:        addDTO.Teacher,
		Room:           addDTO.Room,
	}

	if err := s.repository.AddLesson(ctx, lesson); err != nil {
		return entities.Lesson{}, err
	}

	s.apps.AsyncEvent(user.StudyPlaceInfo.ID, "AddLesson", lesson)

	return lesson, nil
}

func (s *controller) UpdateLesson(ctx context.Context, updateDTO dto2.UpdateLessonDTO, user auth.User) error {
	if err := s.validator.UpdateLesson(updateDTO); err != nil {
		return err
	}

	lesson := entities.Lesson{
		Id:             updateDTO.Id,
		StudyPlaceId:   user.StudyPlaceInfo.ID,
		PrimaryColor:   updateDTO.PrimaryColor,
		SecondaryColor: updateDTO.SecondaryColor,
		EndDate:        updateDTO.EndDate,
		StartDate:      updateDTO.StartDate,
		LessonIndex:    updateDTO.LessonIndex,
		Subject:        updateDTO.Subject,
		Group:          updateDTO.Group,
		Teacher:        updateDTO.Teacher,
		Room:           updateDTO.Room,
		Type:           updateDTO.Type,
		Title:          updateDTO.Title,
		Homework:       updateDTO.Homework,
		Description:    updateDTO.Description,
	}

	err, studyPlace := s.repository.GetStudyPlaceByID(ctx, user.StudyPlaceInfo.ID, false)
	if err != nil {
		return err
	}

	var lessonType general.LessonType
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

	err = s.repository.UpdateLesson(ctx, lesson)

	s.apps.AsyncEvent(user.StudyPlaceInfo.ID, "UpdateLesson", lesson)

	return err
}

func (s *controller) DeleteLesson(ctx context.Context, idHex string, user auth.User) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return errors.Wrap(NotValidParams, "id")
	}

	lesson, err := s.repository.GetLessonByID(ctx, id)
	if err != nil {
		return err
	}

	s.apps.Event(user.StudyPlaceInfo.ID, "RemoveLesson", lesson)

	if err = s.repository.DeleteLesson(ctx, id, user.StudyPlaceInfo.ID); err != nil {
		return err
	}

	return nil
}

func (s *controller) GetLessonsByDateAndID(ctx context.Context, user auth.User, idHex string) ([]entities.Lesson, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, errors.Wrap(NotValidParams, "id")
	}

	if user.StudyPlaceInfo.Role == "group" {
		return s.repository.GetFullLessonsByIDAndDate(ctx, user.Id, id)
	}

	return s.repository.GetFullLessonsByIDAndDate(ctx, user.Id, id)
}

func (s *controller) GetLessonByID(ctx context.Context, user auth.User, idHex string) (entities.Lesson, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return entities.Lesson{}, errors.Wrap(NotValidParams, "id")
	}

	if user.StudyPlaceInfo.Role == "group" {
		lesson, err := s.repository.GetFullLessonByID(ctx, id)
		var marks []journalEntities.Mark
		for _, mark := range lesson.Marks {
			if mark.StudentID == user.Id {
				marks = append(marks, mark)
			}
		}
		lesson.Marks = marks

		if err != nil {
			return entities.Lesson{}, err
		}

		return lesson, nil
	}

	lesson, err := s.repository.GetLessonByID(ctx, id)
	if err != nil {
		return entities.Lesson{}, err
	}

	return lesson, nil
}

func (s *controller) RemoveLessonBetweenDates(ctx context.Context, user auth.User, date1, date2 time.Time) error {
	if !date1.Before(date2) {
		return errors.Wrap(validators.ValidationError, "start time is after end time")
	}

	return s.repository.RemoveLessonBetweenDates(ctx, date1, date2, user.StudyPlaceInfo.ID)
}

func (s *controller) SaveCurrentScheduleAsGeneral(ctx context.Context, user auth.User, role string, roleName string) error {
	startDate, endDate := s.scheduleDated(time.Time{}, time.Time{})
	schedule, err := s.repository.GetSchedule(ctx, user.StudyPlaceInfo.ID, role, roleName, startDate, endDate, false, false)
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
			StudyPlaceId:   user.StudyPlaceInfo.ID,
			EndTime:        lesson.EndDate.Format("15:04"),
			StartTime:      lesson.StartDate.Format("15:04"),
			PrimaryColor:   lesson.PrimaryColor,
			SecondaryColor: lesson.SecondaryColor,
			Subject:        lesson.Subject,
			Group:          lesson.Group,
			Teacher:        lesson.Teacher,
			Room:           lesson.Room,
			DayIndex:       dayIndex,
			WeekIndex:      weekIndex,
		}

		lessons[i] = gLesson
	}

	if err = s.repository.RemoveGeneralLessonsByType(ctx, user.StudyPlaceInfo.ID, role, roleName); err != nil {
		return err
	}

	if err = s.repository.UpdateGeneralSchedule(ctx, lessons); err != nil {
		return err
	}

	return nil
}

func (s *controller) SaveGeneralScheduleAsCurrent(ctx context.Context, user auth.User, date time.Time) error {
	startDayDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	err, studyPlace := s.generalController.GetStudyPlaceByID(ctx, user.StudyPlaceInfo.ID, false)
	if err != nil {
		return err
	}

	_, week := date.ISOWeek()
	weekday := int(date.Weekday()) - 1
	if weekday == -1 {
		weekday = 6
	}

	generalLessons, err := s.repository.GetGeneralLessons(ctx, user.StudyPlaceInfo.ID, week%studyPlace.WeeksCount, weekday)
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
			StudyPlaceId:   user.StudyPlaceInfo.ID,
			PrimaryColor:   generalLesson.PrimaryColor,
			SecondaryColor: generalLesson.SecondaryColor,
			Type:           generalLesson.Type,
			StartDate:      startDate,
			EndDate:        endDate,
			Subject:        generalLesson.Subject,
			Group:          generalLesson.Group,
			Teacher:        generalLesson.Teacher,
			Room:           generalLesson.Room,
			LessonIndex:    generalLesson.LessonIndex,
		}

		lessons[i] = lesson
	}

	if err = s.repository.RemoveLessonBetweenDates(ctx, startDayDate, startDayDate.AddDate(0, 0, 1), user.StudyPlaceInfo.ID); err != nil {
		return err
	}
	return s.repository.AddLessons(ctx, lessons)
}
