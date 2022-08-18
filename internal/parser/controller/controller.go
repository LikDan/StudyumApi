package controller

import (
	"context"
	"errors"
	"studyum/internal/parser/apps"
	"studyum/internal/parser/entities"
	"studyum/internal/parser/repository"
)

type Controller interface {
	Apps() []entities.IApp

	UpdateGeneralSchedule(app entities.IApp)
	Update(ctx context.Context, app entities.IApp)
	GetLastLesson(ctx context.Context, id int) (entities.Lesson, error)
	InsertScheduleTypes(ctx context.Context, types []entities.ScheduleTypeInfo) error
	GetAppByStudyPlaceId(id int) (entities.IApp, error)
}

type controller struct {
	repository repository.ParserRepository

	apps []entities.IApp
}

func NewParserController(repository repository.ParserRepository) Controller {
	return &controller{
		repository: repository,
		apps:       []entities.IApp{&apps.KbpApp},
	}
}

func (c *controller) Apps() []entities.IApp {
	return c.apps
}

func (c *controller) UpdateGeneralSchedule(app entities.IApp) {
	ctx := context.Background()

	var types []entities.ScheduleTypeInfo
	if err := c.repository.GetScheduleTypesToParse(ctx, app.GetName(), types); err != nil {
		return
	}

	for _, type_ := range types {
		lessons := app.GeneralScheduleUpdate(type_)
		_ = c.repository.UpdateGeneralSchedule(ctx, lessons)
	}
}

func (c *controller) Update(ctx context.Context, app entities.IApp) {
	var users []entities.JournalUser
	if err := c.repository.GetUsersToParse(ctx, app.GetName(), users); err != nil {
		return
	}

	for _, user := range users {
		marks := app.JournalUpdate(user, c.repository.GetLessonByDate)

		if err := c.repository.AddMarks(ctx, marks); err != nil {
			continue
		}

		_ = c.repository.UpdateParseJournalUser(ctx, user)
	}

	var types []entities.ScheduleTypeInfo
	if err := c.repository.GetScheduleTypesToParse(ctx, app.GetName(), types); err != nil {
		return
	}

	for _, type_ := range types {
		lessons := app.ScheduleUpdate(type_)
		_ = c.repository.AddLessons(ctx, lessons)
	}
}

func (c *controller) GetLastLesson(ctx context.Context, id int) (entities.Lesson, error) {
	var lesson entities.Lesson
	if err := c.repository.GetLastLesson(ctx, id, lesson); err != nil {
		return entities.Lesson{}, err
	}

	return lesson, nil
}

func (c *controller) InsertScheduleTypes(ctx context.Context, types []entities.ScheduleTypeInfo) error {
	return c.repository.InsertScheduleTypes(ctx, types)
}

func (c *controller) GetAppByStudyPlaceId(id int) (entities.IApp, error) {
	for _, app := range c.apps {
		if app.StudyPlaceId() == id {
			return app, nil
		}
	}

	return nil, errors.New("no app with this id")
}
