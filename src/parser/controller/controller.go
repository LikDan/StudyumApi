package controller

import (
	"context"
	"errors"
	"studyum/src/parser/apps"
	"studyum/src/parser/entities"
	"studyum/src/parser/repository"
)

type Controller struct {
	repository repository.IParserRepository

	apps []entities.IApp
}

func NewParserController(repository repository.IParserRepository) *Controller {
	return &Controller{
		repository: repository,
		apps:       []entities.IApp{&apps.KbpApp},
	}
}

func (c *Controller) Apps() []entities.IApp {
	return c.apps
}

func (c *Controller) UpdateGeneralSchedule(app entities.IApp) {
	ctx := context.Background()

	var types []entities.ScheduleTypeInfo
	if err := c.repository.GetScheduleTypesToParse(ctx, app.GetName(), &types); err != nil {
		return
	}

	for _, type_ := range types {
		lessons := app.GeneralScheduleUpdate(&type_)
		_ = c.repository.UpdateGeneralSchedule(ctx, lessons)
	}
}

func (c *Controller) Update(ctx context.Context, app entities.IApp) {
	var users []entities.JournalUser
	if err := c.repository.GetUsersToParse(ctx, app.GetName(), &users); err != nil {
		return
	}

	for _, user := range users {
		marks := app.JournalUpdate(&user, c.repository.GetLessonByDate)

		if err := c.repository.AddMarks(ctx, marks); err != nil {
			continue
		}

		_ = c.repository.UpdateParseJournalUser(ctx, &user)
	}

	var types []entities.ScheduleTypeInfo
	if err := c.repository.GetScheduleTypesToParse(ctx, app.GetName(), &types); err != nil {
		return
	}

	for _, type_ := range types {
		lessons := app.ScheduleUpdate(&type_)
		_ = c.repository.AddLessons(ctx, lessons)
	}
}

func (c *Controller) GetLastLesson(ctx context.Context, id int) (entities.Lesson, error) {
	var lesson entities.Lesson
	if err := c.repository.GetLastLesson(ctx, id, &lesson); err != nil {
		return entities.Lesson{}, err
	}

	return lesson, nil
}

func (c *Controller) InsertScheduleTypes(ctx context.Context, types []*entities.ScheduleTypeInfo) error {

	if err := c.repository.InsertScheduleTypes(ctx, types); err != nil {
		return err
	}

	return nil
}

func (c *Controller) GetAppByStudyPlaceId(id int) (entities.IApp, error) {
	for _, app := range c.apps {
		if app.StudyPlaceId() == id {
			return app, nil
		}
	}

	return nil, errors.New("no app with this id")
}
