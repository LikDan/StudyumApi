package controller

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/application"
	"studyum/internal/parser/apps"
	"studyum/internal/parser/entities"
	"studyum/internal/parser/repository"
	"studyum/internal/utils"
)

type Controller interface {
	Apps() []application.App

	UpdateGeneralSchedule(app application.App)
	Update(ctx context.Context, app application.App)
	GetLastLesson(ctx context.Context, id int) (entities.Lesson, error)
	InsertScheduleTypes(ctx context.Context, types []entities.ScheduleTypeInfo) error
	GetAppByStudyPlaceId(id int) (application.App, error)
}

type controller struct {
	repository repository.Repository

	apps []application.App
}

func NewParserController(repository repository.Repository) Controller {
	return &controller{
		repository: repository,
		apps:       []application.App{&apps.KbpApp},
	}
}

func (c *controller) Apps() []application.App {
	return c.apps
}

func (c *controller) UpdateGeneralSchedule(app application.App) {
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
				EndTime:      utils.FormatDuration(lessonDTO.Shift.End),
				StartTime:    utils.FormatDuration(lessonDTO.Shift.Start),
				Subject:      lessonDTO.Subject,
				Group:        lessonDTO.Group,
				Teacher:      lessonDTO.Teacher,
				Room:         lessonDTO.Room,
				DayIndex:     lessonDTO.Shift.Date.Day(),
				WeekIndex:    lessonDTO.WeekIndex,
			}

			lessons[i] = lesson
		}

		_ = c.repository.UpdateGeneralSchedule(ctx, lessons)
	}
}

func (c *controller) Update(ctx context.Context, app application.App) {
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
				UserId:       markDTO.UserId,
				LessonId:     lessonID,
				StudyPlaceId: app.StudyPlaceId(),
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
				Id:           primitive.NewObjectID(),
				StudyPlaceId: app.StudyPlaceId(),
				Type:         lessonDTO.Type,
				EndDate:      lessonDTO.Shift.Date.Add(lessonDTO.Shift.End),
				StartDate:    lessonDTO.Shift.Date.Add(lessonDTO.Shift.Start),
				Subject:      lessonDTO.Subject,
				Group:        lessonDTO.Group,
				Teacher:      lessonDTO.Teacher,
				Room:         lessonDTO.Room,
			}

			lessons[i] = lesson
		}

		_ = c.repository.AddLessons(ctx, lessons)
	}
}

func (c *controller) GetLastLesson(ctx context.Context, id int) (entities.Lesson, error) {
	lesson, err := c.repository.GetLastLesson(ctx, id)
	if err != nil {
		return entities.Lesson{}, err
	}

	return lesson, nil
}

func (c *controller) InsertScheduleTypes(ctx context.Context, types []entities.ScheduleTypeInfo) error {
	return c.repository.InsertScheduleTypes(ctx, types)
}

func (c *controller) GetAppByStudyPlaceId(id int) (application.App, error) {
	for _, app := range c.apps {
		if app.StudyPlaceId() == id {
			return app, nil
		}
	}

	return nil, errors.New("no application with this id")
}
