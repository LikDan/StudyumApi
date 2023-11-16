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
	GetSchedule(ctx context.Context, user auth.User, studyPlaceID, role, roleName string, startDate, endDate time.Time, isGeneral bool) (entities.Schedule, error)
	GetUserSchedule(ctx context.Context, user auth.User, startDate, endDate time.Time, isGeneral bool) (entities.Schedule, error)

	GetGeneralSchedule(ctx context.Context, user auth.User, studyPlaceID string, role string, roleName string, startDate, endDate time.Time) (entities.Schedule, error)
	GetGeneralUserSchedule(ctx context.Context, user auth.User, startDate, endDate time.Time) (entities.Schedule, error)

	GetScheduleTypes(ctx context.Context, user auth.User, idHex string) entities.Types

	AddGeneralLessons(ctx context.Context, user auth.User, lessonsDTO []dto2.AddGeneralLessonDTO) ([]entities.GeneralLesson, error)
	AddLessons(ctx context.Context, user auth.User, lessonsDTO []dto2.AddLessonDTO) ([]entities.Lesson, error)

	AddLesson(ctx context.Context, lesson dto2.AddLessonDTO, user auth.User) (entities.Lesson, error)
	GetLessonByID(ctx context.Context, user auth.User, idHex string) (entities.Lesson, error)
	UpdateLesson(ctx context.Context, lessonID string, lesson dto2.UpdateLessonDTO, user auth.User) (entities.Lesson, error)
	DeleteLesson(ctx context.Context, idHex string, user auth.User) error

	GetLessonsByDateAndID(ctx context.Context, user auth.User, idHex string) ([]entities.Lesson, error)

	RemoveLessonBetweenDates(ctx context.Context, user auth.User, date1, date2 time.Time) error

	SaveCurrentScheduleAsGeneral(ctx context.Context, user auth.User, role string, roleName string) error
	SaveGeneralScheduleAsCurrent(ctx context.Context, user auth.User, date time.Time) error

	AddScheduleInfo(ctx context.Context, dto dto2.AddScheduleInfoDTO, user auth.User) (entities.ScheduleInfoEntry, error)

	GetGeneralLessons(ctx context.Context, user auth.User, id string, weekIndex *int, dayIndex *int) ([]entities.GeneralLesson, error)
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

func (s *controller) GetSchedule(ctx context.Context, user auth.User, studyPlaceIDHex, type_, typeName string, startDate, endDate time.Time, isGeneral bool) (entities.Schedule, error) {
	if type_ == "" || typeName == "" {
		return entities.Schedule{}, NotValidParams
	}

	var studyPlaceID primitive.ObjectID
	if user.StudyPlaceInfo != nil {
		studyPlaceID = user.StudyPlaceInfo.ID
	}

	if id, err := primitive.ObjectIDFromHex(studyPlaceIDHex); err == nil {
		studyPlaceID = id
	}

	startDate, endDate = s.scheduleDated(startDate, endDate)

	typeID, err := s.repository.GetTypeID(ctx, studyPlaceID, type_, typeName)
	if err != nil {
		return entities.Schedule{}, err
	}

	return s.repository.GetSchedule(ctx, studyPlaceID, type_, typeName, typeID, startDate, endDate, isGeneral, false)
}

func (s *controller) GetUserSchedule(ctx context.Context, user auth.User, startDate, endDate time.Time, isGeneral bool) (entities.Schedule, error) {
	if user.StudyPlaceInfo.Role == "" || user.StudyPlaceInfo.RoleName == "" {
		return entities.Schedule{}, NotValidParams
	}

	startDate, endDate = s.scheduleDated(startDate, endDate)

	typeID, err := s.repository.GetTypeID(ctx, user.StudyPlaceInfo.ID, user.StudyPlaceInfo.Role, user.StudyPlaceInfo.RoleName)
	if err != nil {
		return entities.Schedule{}, err
	}

	return s.repository.GetSchedule(ctx, user.StudyPlaceInfo.ID, user.StudyPlaceInfo.Role, user.StudyPlaceInfo.RoleName, typeID, startDate, endDate, isGeneral, false)
}

func (s *controller) GetGeneralSchedule(ctx context.Context, user auth.User, studyPlaceIDHex string, type_ string, typeName string, startDate, endDate time.Time) (entities.Schedule, error) {
	if type_ == "" || typeName == "" {
		return entities.Schedule{}, NotValidParams
	}

	studyPlaceID := user.StudyPlaceInfo.ID
	restricted := true
	if id, err := primitive.ObjectIDFromHex(studyPlaceIDHex); err == nil && id != user.StudyPlaceInfo.ID {
		studyPlaceID = id
		restricted = false
	}

	startDate, endDate = s.scheduleDated(startDate, endDate)

	typeID, err := s.repository.GetTypeID(ctx, studyPlaceID, type_, typeName)
	if err != nil {
		return entities.Schedule{}, err
	}

	return s.repository.GetSchedule(ctx, studyPlaceID, type_, typeName, typeID, startDate, endDate, true, !restricted)
}

func (s *controller) GetGeneralUserSchedule(ctx context.Context, user auth.User, startDate, endDate time.Time) (entities.Schedule, error) {
	startDate, endDate = s.scheduleDated(startDate, endDate)

	typeID, err := s.repository.GetTypeID(ctx, user.StudyPlaceInfo.ID, user.StudyPlaceInfo.Role, user.StudyPlaceInfo.RoleName)
	if err != nil {
		return entities.Schedule{}, err
	}
	return s.repository.GetSchedule(ctx, user.StudyPlaceInfo.ID, user.StudyPlaceInfo.Role, user.StudyPlaceInfo.RoleName, typeID, startDate, endDate, true, false)
}

func (s *controller) GetScheduleTypes(ctx context.Context, user auth.User, idHex string) entities.Types {
	studyPlaceID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil && user.StudyPlaceInfo != nil {
		studyPlaceID = user.StudyPlaceInfo.ID
	}

	if studyPlaceID.IsZero() {
		return entities.Types{}
	}

	subjects, err := s.repository.GetScheduleType(ctx, studyPlaceID, "Subjects", "subject")
	if err != nil {
		return entities.Types{}
	}
	teachers, err := s.repository.GetScheduleTeacherType(ctx, studyPlaceID)
	if err != nil {
		return entities.Types{}
	}
	groups, err := s.repository.GetScheduleType(ctx, studyPlaceID, "Groups", "group")
	if err != nil {
		return entities.Types{}
	}
	rooms, err := s.repository.GetScheduleType(ctx, studyPlaceID, "Rooms", "room")
	if err != nil {
		return entities.Types{}
	}

	return entities.Types{
		Subjects: subjects,
		Teachers: teachers,
		Groups:   groups,
		Rooms:    rooms,
	}
}

func (s *controller) AddGeneralLessons(ctx context.Context, user auth.User, lessonsDTO []dto2.AddGeneralLessonDTO) ([]entities.GeneralLesson, error) {
	lessons := make([]entities.GeneralLesson, 0, len(lessonsDTO))
	for _, lessonDTO := range lessonsDTO {
		if err := s.validator.AddGeneralLesson(lessonDTO); err != nil {
			return nil, err
		}

		lesson := entities.GeneralLesson{
			Id:               primitive.NewObjectID(),
			StudyPlaceId:     user.StudyPlaceInfo.ID,
			PrimaryColor:     lessonDTO.PrimaryColor,
			SecondaryColor:   lessonDTO.SecondaryColor,
			StartTimeMinutes: lessonDTO.StartTimeMinutes,
			EndTimeMinutes:   lessonDTO.EndTimeMinutes,
			SubjectID:        lessonDTO.SubjectID,
			GroupID:          lessonDTO.GroupID,
			TeacherID:        lessonDTO.TeacherID,
			RoomID:           lessonDTO.RoomID,
			LessonIndex:      lessonDTO.LessonIndex,
			DayIndex:         lessonDTO.DayIndex,
			WeekIndex:        lessonDTO.WeekIndex,
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
			SubjectID:      lessonDTO.SubjectID,
			GroupID:        lessonDTO.GroupID,
			TeacherID:      lessonDTO.TeacherID,
			RoomID:         lessonDTO.RoomID,
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
		EndDate:        addDTO.EndDate.UTC(),
		StartDate:      addDTO.StartDate.UTC(),
		LessonIndex:    addDTO.LessonIndex,
		SubjectID:      addDTO.SubjectID,
		GroupID:        addDTO.GroupID,
		TeacherID:      addDTO.TeacherID,
		RoomID:         addDTO.RoomID,
	}

	if err := s.repository.AddLesson(ctx, lesson); err != nil {
		return entities.Lesson{}, err
	}

	s.apps.AsyncEvent(user.StudyPlaceInfo.ID, "AddLesson", lesson)

	return s.repository.GetLessonByID(ctx, lesson.Id)
}

func (s *controller) UpdateLesson(ctx context.Context, lessonIDHex string, updateDTO dto2.UpdateLessonDTO, user auth.User) (entities.Lesson, error) {
	if err := s.validator.UpdateLesson(updateDTO); err != nil {
		return entities.Lesson{}, err
	}

	lessonID, err := primitive.ObjectIDFromHex(lessonIDHex)
	if err != nil {
		return entities.Lesson{}, err
	}

	lesson := entities.Lesson{
		Id:             lessonID,
		StudyPlaceId:   user.StudyPlaceInfo.ID,
		PrimaryColor:   updateDTO.PrimaryColor,
		SecondaryColor: updateDTO.SecondaryColor,
		StartDate:      updateDTO.StartDate.UTC(),
		EndDate:        updateDTO.EndDate.UTC(),
		LessonIndex:    updateDTO.LessonIndex,
		SubjectID:      updateDTO.SubjectID,
		GroupID:        updateDTO.GroupID,
		TeacherID:      updateDTO.TeacherID,
		RoomID:         updateDTO.RoomID,
		Type:           updateDTO.Type,
		Title:          updateDTO.Title,
		Homework:       updateDTO.Homework,
		Description:    updateDTO.Description,
	}

	err, studyPlace := s.repository.GetStudyPlaceByID(ctx, user.StudyPlaceInfo.ID, false)
	if err != nil {
		return entities.Lesson{}, err
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
		return entities.Lesson{}, err
	}

	err = s.repository.UpdateLesson(ctx, lesson)

	s.apps.AsyncEvent(user.StudyPlaceInfo.ID, "UpdateLesson", lesson)

	return s.repository.GetLessonByID(ctx, lesson.Id)
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
	schedule, err := s.repository.GetSchedule(ctx, user.StudyPlaceInfo.ID, role, roleName, primitive.NewObjectID(), startDate, endDate, false, false)
	if err != nil {
		return err
	}

	lessons := make([]entities.GeneralLesson, len(schedule.Lessons))
	for i, lesson := range schedule.Lessons {
		_, weekIndex := lesson.StartDate.ISOWeek()
		dayIndex := int(lesson.StartDate.Weekday())

		gLesson := entities.GeneralLesson{
			Id:               primitive.NewObjectID(),
			StudyPlaceId:     user.StudyPlaceInfo.ID,
			EndTimeMinutes:   lesson.EndDate.Hour()*60 + lesson.EndDate.Minute(),
			StartTimeMinutes: lesson.StartDate.Hour()*60 + lesson.StartDate.Minute(),
			PrimaryColor:     lesson.PrimaryColor,
			SecondaryColor:   lesson.SecondaryColor,
			Subject:          lesson.Subject,
			Group:            lesson.Group,
			Teacher:          lesson.Teacher,
			Room:             lesson.Room,
			DayIndex:         dayIndex,
			WeekIndex:        weekIndex,
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
	weekday := int(date.Weekday())

	generalLessons, err := s.repository.GetGeneralLessons(ctx, user.StudyPlaceInfo.ID, week%studyPlace.WeeksCount, weekday)
	if err != nil {
		return err
	}

	lessons := make([]entities.Lesson, len(generalLessons))
	for i, generalLesson := range generalLessons {
		startDate := date.Add(time.Duration(generalLesson.StartTimeMinutes) * time.Minute)
		endDate := date.Add(time.Duration(generalLesson.EndTimeMinutes) * time.Minute)

		lesson := entities.Lesson{
			Id:             primitive.NewObjectID(),
			StudyPlaceId:   user.StudyPlaceInfo.ID,
			PrimaryColor:   generalLesson.PrimaryColor,
			SecondaryColor: generalLesson.SecondaryColor,
			Type:           generalLesson.Type,
			StartDate:      startDate,
			EndDate:        endDate,
			SubjectID:      generalLesson.SubjectID,
			GroupID:        generalLesson.GroupID,
			TeacherID:      generalLesson.TeacherID,
			RoomID:         generalLesson.RoomID,
			LessonIndex:    generalLesson.LessonIndex,
		}

		lessons[i] = lesson
	}

	if err = s.repository.RemoveLessonBetweenDates(ctx, startDayDate, startDayDate.AddDate(0, 0, 1), user.StudyPlaceInfo.ID); err != nil {
		return err
	}
	return s.repository.AddLessons(ctx, lessons)
}

func (s *controller) AddScheduleInfo(ctx context.Context, dto dto2.AddScheduleInfoDTO, user auth.User) (entities.ScheduleInfoEntry, error) {
	entry := entities.ScheduleInfoEntry{
		ID:           primitive.NewObjectID(),
		Date:         dto.Date,
		Status:       dto.Status,
		StudyPlaceId: user.StudyPlaceInfo.ID,
	}

	if err := s.repository.RemoveScheduleInfo(ctx, user.StudyPlaceInfo.ID, dto.Date); err != nil {
		return entities.ScheduleInfoEntry{}, err
	}

	if err := s.repository.AddScheduleInfo(ctx, entry); err != nil {
		return entities.ScheduleInfoEntry{}, err
	}

	return entry, nil
}

func (s *controller) GetGeneralLessons(ctx context.Context, user auth.User, _ string, weekIndex *int, dayIndex *int) ([]entities.GeneralLesson, error) {
	if user.StudyPlaceInfo == nil {
		return nil, errors.New("Not authenticated")
	}

	if weekIndex == nil {
		i := 0
		weekIndex = &i
	}

	if dayIndex == nil {
		return s.repository.GetAllGeneralLessons(ctx, user.StudyPlaceInfo.ID)
	}

	return s.repository.GetGeneralLessons(ctx, user.StudyPlaceInfo.ID, *weekIndex, *dayIndex)
}
