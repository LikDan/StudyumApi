package controller

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/apps"
	"studyum/internal/parser/apps/kbp"
	"studyum/internal/parser/dto"
	"studyum/internal/parser/entities"
	"studyum/internal/parser/repository"
	"studyum/pkg/datetime"
	"time"
)

type Controller interface {
	Apps() []apps.App

	UpdateGeneralSchedule(app apps.App)
	Update(ctx context.Context, app apps.App)
	GetLastUpdatedDate(ctx context.Context, id int) (error, time.Time)
	InsertScheduleTypes(ctx context.Context, types []dto.ScheduleTypeInfoDTO) error
	GetAppByStudyPlaceId(id int) (apps.App, error)

	AddMark(ctx context.Context, markDTO dto.Mark)
	EditMark(ctx context.Context, markDTO dto.Mark)
	DeleteMark(ctx context.Context, markDTO dto.Mark)

	AddLesson(ctx context.Context, lessonDTO dto.Lesson)
	EditLesson(ctx context.Context, lessonDTO dto.Lesson)
	DeleteLesson(ctx context.Context, lessonDTO dto.Lesson)
}

type controller struct {
	repository repository.Repository

	apps []apps.App
}

func NewParserController(repository repository.Repository) Controller {
	return &controller{
		repository: repository,
		apps:       []apps.App{kbp.NewApp()},
	}
}

func (c *controller) Apps() []apps.App {
	return c.apps
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

func (c *controller) Update(ctx context.Context, app apps.App) {
	var users []entities.JournalUser
	if _, err := c.repository.GetUsersToParse(ctx, app.GetName()); err != nil {
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

		if err := c.repository.AddMarks(ctx, marks); err != nil {
			continue
		}

		_ = c.repository.UpdateParseJournalUser(ctx, user)
	}

	types, err := c.repository.GetScheduleTypesToParse(ctx, app.GetName())
	if err != nil {
		return
	}

	for _, typeInfo := range types {
		lessonsDTO := app.ScheduleUpdate(typeInfo)

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

func (c *controller) InsertScheduleTypes(ctx context.Context, dto []dto.ScheduleTypeInfoDTO) error {
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

func (c *controller) GetAppByStudyPlaceId(id int) (apps.App, error) {
	for _, app := range c.apps {
		if app.StudyPlaceId() == id {
			return app, nil
		}
	}

	return nil, errors.New("no application with this id")
}

func (c *controller) AddMark(ctx context.Context, markDTO dto.Mark) {
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
	_ = c.repository.UpdateMarkParsedInfoByID(ctx, mark.Id, info)
}

func (c *controller) EditMark(ctx context.Context, markDTO dto.Mark) {
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
	_ = c.repository.UpdateMarkParsedInfoByID(ctx, mark.Id, info)
}

func (c *controller) DeleteMark(ctx context.Context, markDTO dto.Mark) {
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

	app.OnMarkDelete(ctx, mark, lesson)
}

func (c *controller) AddLesson(ctx context.Context, lessonDTO dto.Lesson) {
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
	_ = c.repository.UpdateLessonParsedInfoByID(ctx, lesson.Id, info)
}

func (c *controller) EditLesson(ctx context.Context, lessonDTO dto.Lesson) {
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
	_ = c.repository.UpdateLessonParsedInfoByID(ctx, lesson.Id, info)
}

func (c *controller) DeleteLesson(ctx context.Context, lessonDTO dto.Lesson) {
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

func (c *controller) GetLastUpdatedDate(ctx context.Context, id int) (error, time.Time) {
	return c.repository.GetLastUpdatedDate(ctx, id)
}
