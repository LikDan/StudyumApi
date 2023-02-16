package marks

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/apps/apps/kbp/shared"
	"studyum/internal/apps/entities"
	appShared "studyum/internal/apps/shared"
	journalEntities "studyum/internal/journal/entities"
)

type controller struct {
	repository Repository

	shared appShared.Shared
	auth   shared.AuthRepository
}

func NewController(repository Repository, shared appShared.Shared, auth shared.AuthRepository) entities.MarksManageInterface {
	return &controller{repository: repository, shared: shared, auth: auth}
}

func (c *controller) authAndGetIDs(ctx context.Context, lessonID, studentID primitive.ObjectID) (string, Mark, error) {
	lesson, err := c.shared.GetLessonByID(ctx, lessonID)
	if err != nil {
		return "", Mark{}, err
	}

	lessonIDString, ok := lesson.Data["lessonID"]
	if !ok {
		return "", Mark{}, errors.New("no data")
	}

	user, err := c.shared.GetUserByID(ctx, studentID)
	if err != nil {
		return "", Mark{}, err
	}

	userID, ok := user.Data["userID"]
	if !ok {
		return "", Mark{}, errors.New("no data")
	}

	mark := Mark{
		LessonID:  lessonIDString.(string),
		StudentID: userID.(string),
	}

	token := c.auth.Auth(ctx)
	if token == "" {
		return "", Mark{}, errors.New("bad credentials")
	}

	return token, mark, nil
}

func (c *controller) AddMark(ctx context.Context, _ appShared.Data, sMark journalEntities.Mark) appShared.Data {
	token, mark, err := c.authAndGetIDs(ctx, sMark.LessonID, sMark.StudentID)
	if err != nil {
		return nil
	}

	mark.Value = sMark.Mark

	id, err := c.repository.AddMark(ctx, token, mark)
	if err != nil {
		return nil
	}

	return appShared.Data{"markID": id}
}

func (c *controller) UpdateMark(ctx context.Context, data appShared.Data, sMark journalEntities.Mark) appShared.Data {
	markID, ok := data["markID"]
	if !ok {
		return nil
	}

	token, mark, err := c.authAndGetIDs(ctx, sMark.LessonID, sMark.StudentID)
	if err != nil {
		return nil
	}

	mark.Value = sMark.Mark
	mark.MarkID = markID.(string)

	id, err := c.repository.UpdateMark(ctx, token, mark)
	if err != nil {
		return nil
	}

	return appShared.Data{"markID": id}
}

func (c *controller) RemoveMark(ctx context.Context, data appShared.Data, sMark journalEntities.Mark) {
	markID, ok := data["markID"]
	if !ok {
		return
	}

	token, mark, err := c.authAndGetIDs(ctx, sMark.LessonID, sMark.StudentID)
	if err != nil {
		return
	}

	mark.MarkID = markID.(string)

	_, _ = c.repository.DeleteMark(ctx, token, mark)
}
