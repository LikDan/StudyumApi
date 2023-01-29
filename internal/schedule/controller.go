package schedule

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	auth "studyum/internal/auth/entities"
	"studyum/internal/general/controllers"
	general "studyum/internal/general/entities"
	"studyum/internal/global"
	"studyum/internal/journal/entities"
	"studyum/internal/parser/dto"
	parser "studyum/internal/parser/handler"
	"time"
)

type Controller interface {
	GetSchedule(ctx context.Context, studyPlaceID string, type_ string, typeName string, user auth.User) (Schedule, error)
	GetUserSchedule(ctx context.Context, user auth.User) (Schedule, error)

	GetGeneralSchedule(ctx context.Context, studyPlaceID string, type_ string, typeName string, user auth.User) (Schedule, error)
	GetGeneralUserSchedule(ctx context.Context, user auth.User) (Schedule, error)

	GetScheduleTypes(ctx context.Context, user auth.User, idHex string) Types

	AddGeneralLessons(ctx context.Context, user auth.User, lessonsDTO []AddGeneralLessonDTO) ([]GeneralLesson, error)
	AddLessons(ctx context.Context, user auth.User, lessonsDTO []AddLessonDTO) ([]Lesson, error)

	AddLesson(ctx context.Context, lesson AddLessonDTO, user auth.User) (Lesson, error)
	GetLessonByID(ctx context.Context, user auth.User, idHex string) (Lesson, error)
	UpdateLesson(ctx context.Context, lesson UpdateLessonDTO, user auth.User) error
	DeleteLesson(ctx context.Context, idHex string, user auth.User) error

	GetLessonsByDateAndID(ctx context.Context, user auth.User, idHex string) ([]Lesson, error)

	RemoveLessonBetweenDates(ctx context.Context, user auth.User, date1, date2 time.Time) error

	SaveCurrentScheduleAsGeneral(ctx context.Context, user auth.User, type_ string, typeName string) error
	SaveGeneralScheduleAsCurrent(ctx context.Context, user auth.User, date time.Time) error
}

type controller struct {
	parser    parser.Handler
	validator Validator

	repository        Repository
	generalController controllers.Controller
}

func NewScheduleController(parser parser.Handler, validator Validator, repository Repository, generalController controllers.Controller) Controller {
	return &controller{parser: parser, validator: validator, repository: repository, generalController: generalController}
}

func (s *controller) GetSchedule(ctx context.Context, studyPlaceIDHex string, type_ string, typeName string, user auth.User) (Schedule, error) {
	if type_ == "" || typeName == "" {
		return Schedule{}, global.NotValidParams
	}

	studyPlaceID := user.StudyPlaceID
	restricted := true
	if id, err := primitive.ObjectIDFromHex(studyPlaceIDHex); err == nil && id != user.StudyPlaceID {
		studyPlaceID = id
		restricted = false
	}

	return s.repository.GetSchedule(ctx, studyPlaceID, type_, typeName, false, !restricted)
}

func (s *controller) GetUserSchedule(ctx context.Context, user auth.User) (Schedule, error) {
	return s.repository.GetSchedule(ctx, user.StudyPlaceID, user.Type, user.TypeName, false, false)
}

func (s *controller) GetGeneralSchedule(ctx context.Context, studyPlaceIDHex string, type_ string, typeName string, user auth.User) (Schedule, error) {
	if type_ == "" || typeName == "" {
		return Schedule{}, global.NotValidParams
	}

	studyPlaceID := user.StudyPlaceID
	restricted := true
	if id, err := primitive.ObjectIDFromHex(studyPlaceIDHex); err == nil && id != user.StudyPlaceID {
		studyPlaceID = id
		restricted = false
	}

	return s.repository.GetSchedule(ctx, studyPlaceID, type_, typeName, true, !restricted)
}

func (s *controller) GetGeneralUserSchedule(ctx context.Context, user auth.User) (Schedule, error) {
	return s.repository.GetSchedule(ctx, user.StudyPlaceID, user.Type, user.TypeName, true, false)
}

func (s *controller) GetScheduleTypes(ctx context.Context, user auth.User, idHex string) Types {
	studyPlaceID := user.StudyPlaceID
	if id, err := primitive.ObjectIDFromHex(idHex); err == nil && user.StudyPlaceID != id {
		if err, _ = s.repository.GetStudyPlaceByID(ctx, id, false); err != nil {
			return Types{}
		}

		studyPlaceID = id
	}

	return Types{
		Groups:   s.repository.GetScheduleType(ctx, studyPlaceID, "group"),
		Teachers: s.repository.GetScheduleType(ctx, studyPlaceID, "teacher"),
		Subjects: s.repository.GetScheduleType(ctx, studyPlaceID, "subject"),
		Rooms:    s.repository.GetScheduleType(ctx, studyPlaceID, "room"),
	}
}

func (s *controller) AddGeneralLessons(ctx context.Context, user auth.User, lessonsDTO []AddGeneralLessonDTO) ([]GeneralLesson, error) {
	lessons := make([]GeneralLesson, 0, len(lessonsDTO))
	for _, lessonDTO := range lessonsDTO {
		if err := s.validator.AddGeneralLesson(lessonDTO); err != nil {
			return nil, err
		}

		lesson := GeneralLesson{
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

func (s *controller) AddLessons(ctx context.Context, user auth.User, lessonsDTO []AddLessonDTO) ([]Lesson, error) {
	lessons := make([]Lesson, 0, len(lessonsDTO))
	for _, lessonDTO := range lessonsDTO {
		if err := s.validator.AddLesson(lessonDTO); err != nil {
			return nil, err
		}

		lesson := Lesson{
			Id:             primitive.NewObjectID(),
			StudyPlaceId:   user.StudyPlaceID,
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

func (s *controller) AddLesson(ctx context.Context, addDTO AddLessonDTO, user auth.User) (Lesson, error) {
	if err := s.validator.AddLesson(addDTO); err != nil {
		return Lesson{}, err
	}

	lesson := Lesson{
		StudyPlaceId:   user.StudyPlaceID,
		PrimaryColor:   addDTO.PrimaryColor,
		SecondaryColor: addDTO.SecondaryColor,
		EndDate:        addDTO.EndDate,
		StartDate:      addDTO.StartDate,
		LessonIndex:    addDTO.LessonIndex,
		Subject:        addDTO.Subject,
		Group:          addDTO.Group,
		Teacher:        addDTO.Teacher,
		Room:           addDTO.Room,
	}

	id, err := s.repository.AddLesson(ctx, lesson)
	if err != nil {
		return Lesson{}, err
	}

	lesson.Id = id

	lessonDTO := dto.LessonDTO{
		Id:             lesson.Id,
		StudyPlaceId:   lesson.StudyPlaceId,
		PrimaryColor:   lesson.PrimaryColor,
		SecondaryColor: lesson.SecondaryColor,
		EndDate:        lesson.EndDate,
		StartDate:      lesson.StartDate,
		Subject:        lesson.Subject,
		Group:          lesson.Group,
		Teacher:        lesson.Teacher,
		Room:           lesson.Room,
		ParsedInfo:     lesson.ParsedInfo,
	}
	go s.parser.AddLesson(lessonDTO)

	return lesson, err
}

func (s *controller) UpdateLesson(ctx context.Context, updateDTO UpdateLessonDTO, user auth.User) error {
	if err := s.validator.UpdateLesson(updateDTO); err != nil {
		return err
	}

	lesson := Lesson{
		Id:             updateDTO.Id,
		StudyPlaceId:   user.StudyPlaceID,
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

	lessonDTO := dto.LessonDTO{
		Id:             lesson.Id,
		StudyPlaceId:   lesson.StudyPlaceId,
		PrimaryColor:   lesson.PrimaryColor,
		SecondaryColor: lesson.SecondaryColor,
		EndDate:        lesson.EndDate,
		StartDate:      lesson.StartDate,
		Subject:        lesson.Subject,
		Group:          lesson.Group,
		Teacher:        lesson.Teacher,
		Room:           lesson.Room,
		ParsedInfo:     lesson.ParsedInfo,
	}
	go s.parser.EditLesson(lessonDTO)

	err, studyPlace := s.repository.GetStudyPlaceByID(ctx, user.StudyPlaceID, false)
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

	return s.repository.UpdateLesson(ctx, lesson)
}

func (s *controller) DeleteLesson(ctx context.Context, idHex string, user auth.User) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return errors.Wrap(global.NotValidParams, "id")
	}

	lesson, err := s.repository.FindAndDeleteLesson(ctx, id, user.StudyPlaceID)
	if err != nil {
		return err
	}

	lessonDTO := dto.LessonDTO{
		Id:             lesson.Id,
		StudyPlaceId:   lesson.StudyPlaceId,
		PrimaryColor:   lesson.PrimaryColor,
		SecondaryColor: lesson.SecondaryColor,
		EndDate:        lesson.EndDate,
		StartDate:      lesson.StartDate,
		Subject:        lesson.Subject,
		Group:          lesson.Group,
		Teacher:        lesson.Teacher,
		Room:           lesson.Room,
		ParsedInfo:     lesson.ParsedInfo,
	}
	go s.parser.DeleteLesson(lessonDTO)

	return nil
}

func (s *controller) GetLessonsByDateAndID(ctx context.Context, user auth.User, idHex string) ([]Lesson, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, errors.Wrap(global.NotValidParams, "id")
	}

	if user.Type == "group" {
		return s.repository.GetFullLessonsByIDAndDate(ctx, user.Id, id)
	}

	return s.repository.GetFullLessonsByIDAndDate(ctx, user.Id, id)
}

func (s *controller) GetLessonByID(ctx context.Context, user auth.User, idHex string) (Lesson, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return Lesson{}, errors.Wrap(global.NotValidParams, "id")
	}

	if user.Type == "group" {
		lesson, err := s.repository.GetFullLessonByID(ctx, id)
		var marks []entities.Mark
		for _, mark := range lesson.Marks {
			if mark.StudentID == user.Id {
				marks = append(marks, mark)
			}
		}
		lesson.Marks = marks

		if err != nil {
			return Lesson{}, err
		}

		return lesson, nil
	}

	lesson, err := s.repository.GetLessonByID(ctx, id)
	if err != nil {
		return Lesson{}, err
	}

	return lesson, nil
}

func (s *controller) RemoveLessonBetweenDates(ctx context.Context, user auth.User, date1, date2 time.Time) error {
	if !date1.Before(date2) {
		return errors.Wrap(global.ValidationError, "start time is after end time")
	}

	return s.repository.RemoveLessonBetweenDates(ctx, date1, date2, user.StudyPlaceID)
}

func (s *controller) SaveCurrentScheduleAsGeneral(ctx context.Context, user auth.User, type_ string, typeName string) error {
	schedule, err := s.repository.GetSchedule(ctx, user.StudyPlaceID, type_, typeName, false, false)
	if err != nil {
		return err
	}

	lessons := make([]GeneralLesson, len(schedule.Lessons))
	for i, lesson := range schedule.Lessons {
		_, weekIndex := lesson.StartDate.ISOWeek()
		dayIndex := int(lesson.StartDate.Weekday()) - 1
		if dayIndex == -1 {
			dayIndex = 6
		}

		gLesson := GeneralLesson{
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

func (s *controller) SaveGeneralScheduleAsCurrent(ctx context.Context, user auth.User, date time.Time) error {
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

	lessons := make([]Lesson, len(generalLessons))
	for i, generalLesson := range generalLessons {
		startDate, err := time.Parse("2006-01-02T15:04", date.Format("2006-01-02T")+generalLesson.StartTime)
		if err != nil {
			return err
		}

		endDate, err := time.Parse("2006-01-02T15:04", date.Format("2006-01-02T")+generalLesson.EndTime)
		if err != nil {
			return err
		}

		lesson := Lesson{
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
