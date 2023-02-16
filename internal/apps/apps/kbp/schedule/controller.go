package schedule

import (
	"context"
	"studyum/internal/apps/apps/kbp/shared"
	"studyum/internal/apps/entities"
	appShared "studyum/internal/apps/shared"
	scheduleEntities "studyum/internal/schedule/entities"
)

type controller struct {
	repository Repository
	mongo      MongoRepository

	shared appShared.Shared
	auth   shared.AuthRepository
}

func NewController(repository Repository, mongo MongoRepository, shared appShared.Shared, auth shared.AuthRepository) entities.LessonsManageInterface {
	return &controller{repository: repository, mongo: mongo, shared: shared, auth: auth}
}

func (c *controller) fulfill(ctx context.Context, sLesson scheduleEntities.Lesson) (Lesson, error) {
	groupID, err := c.mongo.GetGroupID(ctx, sLesson.Group)
	if err != nil {
		return Lesson{}, nil
	}

	subjectID, err := c.mongo.GetSubjectID(ctx, sLesson.Subject)
	if err != nil {
		return Lesson{}, nil
	}

	lesson := Lesson{
		Date:        sLesson.StartDate.Format("2006-01-02"),
		Description: sLesson.Description,
		PairType:    -1,
		SubjectID:   subjectID,
		GroupID:     groupID,
	}

	switch sLesson.Type {
	case "Лекция":
		lesson.PairType = 0
	case "Практика":
		lesson.PairType = 1
	case "Лабораторная":
		lesson.PairType = 2
	}

	return lesson, nil
}

func (c *controller) AddLesson(ctx context.Context, _ appShared.Data, sLesson scheduleEntities.Lesson) appShared.Data {
	lesson, err := c.fulfill(ctx, sLesson)
	if err != nil {
		return nil
	}

	token := c.auth.Auth(ctx)
	id, err := c.repository.AddLesson(ctx, token, lesson)
	if err != nil {
		return nil
	}

	return appShared.Data{"lessonID": id}
}

func (c *controller) UpdateLesson(ctx context.Context, data appShared.Data, sLesson scheduleEntities.Lesson) appShared.Data {
	lessonID, ok := data["lessonID"]
	if !ok {
		return nil
	}

	lesson, err := c.fulfill(ctx, sLesson)
	if err != nil {
		return nil
	}

	lesson.ID = lessonID.(string)

	token := c.auth.Auth(ctx)
	if _, err = c.repository.AddLesson(ctx, token, lesson); err != nil {
		return nil
	}

	return data
}

func (c *controller) RemoveLesson(context.Context, appShared.Data, scheduleEntities.Lesson) {
}
