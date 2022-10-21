package controller

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/appDTO"
	"studyum/internal/parser/apps"
	"studyum/internal/parser/apps/kbp"
	"studyum/internal/parser/dto"
	"studyum/internal/parser/entities"
	"studyum/internal/parser/repository"
	"studyum/pkg/datetime"
	"studyum/pkg/firebase"
	"studyum/pkg/slicetools"
	"time"
)

type Controller interface {
	Apps() []apps.App

	UpdateGeneralSchedule(app apps.App)
	UpdateSchedule(ctx context.Context, app apps.App)
	UpdateJournal(ctx context.Context, app apps.App)
	GetLastUpdatedDate(ctx context.Context, id primitive.ObjectID) (error, time.Time)
	InsertScheduleTypes(ctx context.Context, types []appDTO.ScheduleTypeInfoDTO) error
	GetAppByStudyPlaceId(id primitive.ObjectID) (apps.App, error)

	AddMark(ctx context.Context, markDTO dto.MarkDTO)
	EditMark(ctx context.Context, markDTO dto.MarkDTO)
	DeleteMark(ctx context.Context, markDTO primitive.ObjectID, id primitive.ObjectID)

	AddLesson(ctx context.Context, lessonDTO dto.LessonDTO)
	EditLesson(ctx context.Context, lessonDTO dto.LessonDTO)
	DeleteLesson(ctx context.Context, lessonDTO dto.LessonDTO)

	GetSignUpDataByCode(ctx context.Context, code string) (entities.SignUpCode, error)
}

type controller struct {
	repository repository.Repository

	firebase firebase.Firebase

	apps []apps.App
}

func NewParserController(repository repository.Repository, firebase firebase.Firebase) Controller {
	return &controller{
		repository: repository,
		firebase:   firebase,
		apps:       []apps.App{kbp.NewApp()},
	}
}

func (c *controller) Apps() []apps.App {
	return c.apps
}

func (c *controller) SendMarkNotification(ctx context.Context, token string, title string, mark entities.Mark) (string, error) {
	lesson, err := c.repository.GetLessonByID(ctx, mark.LessonId)
	if err != nil {
		return "", err
	}

	date := lesson.StartDate.Format("January 2 (Monday) 3:4 PM - ") + lesson.EndDate.Format("3:4 PM")
	return c.firebase.SendNotification(ctx, token, "journal", title, mark.Mark+" - "+lesson.Subject+" for "+date, "")
}

func (c *controller) UpdateGeneralSchedule(app apps.App) {
	ctx := context.Background()

	types, err := c.repository.GetScheduleTypesToParse(ctx, app.GetName())
	if err != nil {
		return
	}

	for _, typeInfo := range types {
		lessonsDTO := app.GeneralScheduleUpdate(typeInfo)
		lessons := make([]entities.GeneralLesson, len(lessonsDTO))
		for i, lessonDTO := range lessonsDTO {
			lesson := entities.GeneralLesson{
				Id:           primitive.NewObjectID(),
				StudyPlaceId: app.StudyPlaceId(),
				EndTime:      datetime.FormatDuration(lessonDTO.Shift.End),
				StartTime:    datetime.FormatDuration(lessonDTO.Shift.Start),
				Subject:      lessonDTO.Subject,
				Group:        lessonDTO.Group,
				Teacher:      lessonDTO.Teacher,
				Room:         lessonDTO.Room,
				DayIndex:     lessonDTO.Shift.Date.Day(),
				WeekIndex:    lessonDTO.WeekIndex,
				ParsedInfo:   lessonDTO.ParsedInfo,
			}

			lessons[i] = lesson
		}

		_ = c.repository.UpdateGeneralSchedule(ctx, lessons)
	}
}

func (c *controller) UpdateSchedule(ctx context.Context, app apps.App) {
	types, err := c.repository.GetScheduleTypesToParse(ctx, app.GetName())
	if err != nil {
		return
	}

	send := true
	for _, typeInfo := range types {
		lessonsDTO := app.ScheduleUpdate(typeInfo)
		if len(lessonsDTO) != 0 && send {
			send = false

			notification, err := c.firebase.SendNotification(ctx, "", "schedule", "Schedule was updated", "Schedule was updated", "")
			if err != nil {
				logrus.Error("error sending notification: " + err.Error())
			} else {
				logrus.Info("sending notification -> " + notification)
			}
		}

		lessons := make([]entities.Lesson, len(lessonsDTO))
		for i, lessonDTO := range lessonsDTO {
			lesson := entities.Lesson{
				Id:             primitive.NewObjectID(),
				StudyPlaceId:   app.StudyPlaceId(),
				PrimaryColor:   lessonDTO.PrimaryColor,
				SecondaryColor: lessonDTO.SecondaryColor,
				EndDate:        lessonDTO.Shift.Date.Add(lessonDTO.Shift.End),
				StartDate:      lessonDTO.Shift.Date.Add(lessonDTO.Shift.Start),
				Subject:        lessonDTO.Subject,
				Group:          lessonDTO.Group,
				Teacher:        lessonDTO.Teacher,
				Room:           lessonDTO.Room,
				ParsedInfo:     lessonDTO.ParsedInfo,
			}

			lessons[i] = lesson
		}

		_ = c.repository.AddLessons(ctx, lessons)
	}
}

func (c *controller) UpdateJournal(ctx context.Context, app apps.App) {
	users, err := c.repository.GetUsersToParse(ctx, app.GetName())
	if err != nil {
		return
	}

	for _, user := range users {
		marksDTO := app.JournalUpdate(user)

		marks := make([]entities.Mark, 0, len(marksDTO))
		for _, markDTO := range marksDTO {
			lessonID, err := c.repository.GetLessonIDByDateNameAndGroup(ctx, markDTO.LessonDate, markDTO.Subject, markDTO.Group)
			if err != nil {
				continue
			}

			mark := entities.Mark{
				Id:           primitive.NewObjectID(),
				Mark:         markDTO.Mark,
				StudentID:    markDTO.StudentID,
				LessonId:     lessonID,
				StudyPlaceId: app.StudyPlaceId(),
				ParsedInfo:   markDTO.ParsedInfo,
			}

			marks = append(marks, mark)
		}

		err, existedMarks := c.repository.GetMarks(ctx, user.ID)
		if err != nil {
			continue
		}

		existedMarks, marks = slicetools.RemoveSameFunc(existedMarks, marks, func(m1, m2 entities.Mark) bool {
			return m1.Mark == m2.Mark && m1.LessonId == m2.LessonId && m2.StudentID == m2.StudentID
		})

		if len(existedMarks) == 0 && len(marks) == 0 {
			continue
		}

		err, mainUser := c.repository.GetUserById(ctx, user.ID)
		if err != nil {
			continue
		}

		for _, mark := range existedMarks {
			notification, err := c.SendMarkNotification(ctx, "Mark removed", mainUser.FirebaseToken, mark)
			if err != nil {
				logrus.Error("error sending notification: " + err.Error())
			}
			logrus.Info("sending notification -> " + notification)
		}

		for _, mark := range marks {
			notification, err := c.SendMarkNotification(ctx, "Mark added", mainUser.FirebaseToken, mark)
			if err != nil {
				logrus.Error("error sending notification: " + err.Error())
			}
			logrus.Info("sending notification -> " + notification)
		}

		if err = c.repository.DeleteMarks(ctx, existedMarks); err != nil {
			continue
		}

		if err = c.repository.AddMarks(ctx, marks); err != nil {
			continue
		}

		_ = c.repository.UpdateParseJournalUser(ctx, user)
	}
}

func (c *controller) InsertScheduleTypes(ctx context.Context, dto []appDTO.ScheduleTypeInfoDTO) error {
	types := make([]entities.ScheduleTypeInfo, len(dto))
	for i, infoDTO := range dto {
		types[i] = entities.ScheduleTypeInfo{
			Id:            primitive.NewObjectID(),
			ParserAppName: infoDTO.ParserAppName,
			Group:         infoDTO.Group,
			Url:           infoDTO.Url,
		}
	}

	return c.repository.InsertScheduleTypes(ctx, types)
}

func (c *controller) GetAppByStudyPlaceId(id primitive.ObjectID) (apps.App, error) {
	for _, app := range c.apps {
		if app.StudyPlaceId() == id {
			return app, nil
		}
	}

	return nil, errors.New("no application with this id")
}

func (c *controller) AddMark(ctx context.Context, markDTO dto.MarkDTO) {
	app, err := c.GetAppByStudyPlaceId(markDTO.StudyPlaceId)
	if err != nil {
		return
	}

	mark := entities.Mark{
		Id:           markDTO.Id,
		Mark:         markDTO.Mark,
		StudentID:    markDTO.StudentID,
		LessonId:     markDTO.LessonId,
		StudyPlaceId: markDTO.StudyPlaceId,
	}

	lesson, err := c.repository.GetLessonByID(ctx, markDTO.LessonId)
	if err != nil {
		return
	}

	info := app.OnMarkAdd(ctx, mark, lesson)
	_ = c.repository.UpdateMarkParsedInfoByID(ctx, mark.Id, entities.ParsedInfoType(info))
}

func (c *controller) EditMark(ctx context.Context, markDTO dto.MarkDTO) {
	app, err := c.GetAppByStudyPlaceId(markDTO.StudyPlaceId)
	if err != nil {
		return
	}

	mark := entities.Mark{
		Id:           markDTO.Id,
		Mark:         markDTO.Mark,
		StudentID:    markDTO.StudentID,
		LessonId:     markDTO.LessonId,
		StudyPlaceId: markDTO.StudyPlaceId,
	}

	lesson, err := c.repository.GetLessonByID(ctx, markDTO.LessonId)
	if err != nil {
		return
	}

	info := app.OnMarkEdit(ctx, mark, lesson)
	_ = c.repository.UpdateMarkParsedInfoByID(ctx, mark.Id, entities.ParsedInfoType(info))
}

func (c *controller) DeleteMark(ctx context.Context, id primitive.ObjectID, studyPlaceID primitive.ObjectID) {
	app, err := c.GetAppByStudyPlaceId(studyPlaceID)
	if err != nil {
		return
	}

	app.OnMarkDelete(ctx, id)
}

func (c *controller) AddLesson(ctx context.Context, lessonDTO dto.LessonDTO) {
	app, err := c.GetAppByStudyPlaceId(lessonDTO.StudyPlaceId)
	if err != nil {
		return
	}

	lesson := entities.Lesson{
		Id:             lessonDTO.Id,
		StudyPlaceId:   lessonDTO.StudyPlaceId,
		PrimaryColor:   lessonDTO.PrimaryColor,
		SecondaryColor: lessonDTO.SecondaryColor,
		EndDate:        lessonDTO.EndDate,
		StartDate:      lessonDTO.StartDate,
		Subject:        lessonDTO.Subject,
		Group:          lessonDTO.Group,
		Teacher:        lessonDTO.Teacher,
		Room:           lessonDTO.Room,
	}

	info := app.OnLessonAdd(ctx, lesson)
	_ = c.repository.UpdateLessonParsedInfoByID(ctx, lesson.Id, entities.ParsedInfoType(info))
}

func (c *controller) EditLesson(ctx context.Context, lessonDTO dto.LessonDTO) {
	app, err := c.GetAppByStudyPlaceId(lessonDTO.StudyPlaceId)
	if err != nil {
		return
	}

	lesson := entities.Lesson{
		Id:             lessonDTO.Id,
		StudyPlaceId:   lessonDTO.StudyPlaceId,
		PrimaryColor:   lessonDTO.PrimaryColor,
		SecondaryColor: lessonDTO.SecondaryColor,
		EndDate:        lessonDTO.EndDate,
		StartDate:      lessonDTO.StartDate,
		Subject:        lessonDTO.Subject,
		Group:          lessonDTO.Group,
		Teacher:        lessonDTO.Teacher,
		Room:           lessonDTO.Room,
	}

	info := app.OnLessonEdit(ctx, lesson)
	_ = c.repository.UpdateLessonParsedInfoByID(ctx, lesson.Id, entities.ParsedInfoType(info))
}

func (c *controller) DeleteLesson(ctx context.Context, lessonDTO dto.LessonDTO) {
	app, err := c.GetAppByStudyPlaceId(lessonDTO.StudyPlaceId)
	if err != nil {
		return
	}

	lesson := entities.Lesson{
		Id:             lessonDTO.Id,
		StudyPlaceId:   lessonDTO.StudyPlaceId,
		PrimaryColor:   lessonDTO.PrimaryColor,
		SecondaryColor: lessonDTO.SecondaryColor,
		EndDate:        lessonDTO.EndDate,
		StartDate:      lessonDTO.StartDate,
		Subject:        lessonDTO.Subject,
		Group:          lessonDTO.Group,
		Teacher:        lessonDTO.Teacher,
		Room:           lessonDTO.Room,
	}

	app.OnLessonDelete(ctx, lesson)
}

func (c *controller) GetLastUpdatedDate(ctx context.Context, id primitive.ObjectID) (error, time.Time) {
	return c.repository.GetLastUpdatedDate(ctx, id)
}

func (c *controller) GetSignUpDataByCode(ctx context.Context, code string) (entities.SignUpCode, error) {
	for _, app := range c.apps {
		codeDTO, err := app.GetSignUpDataByCode(ctx, code)
		if err != nil {
			continue
		}

		return entities.SignUpCode{
			Code:         codeDTO.Code,
			Name:         codeDTO.Name,
			StudyPlaceID: app.StudyPlaceId(),
			Type:         codeDTO.Type,
			Typename:     codeDTO.Typename,
		}, nil
	}

	return entities.SignUpCode{}, errors.New("bad code")
}
