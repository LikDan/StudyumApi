package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
	"studyum/internal/repositories"
	"studyum/internal/utils"
)

var (
	NotValidParams = errors.New("not valid params")
	NoPermission   = errors.New("no permission")
)

type JournalController interface {
	GetJournalAvailableOptions(ctx context.Context, user entities.User) ([]entities.JournalAvailableOption, error)

	GetJournal(ctx context.Context, group string, subject string, teacher string, user entities.User) (entities.Journal, error)
	GetUserJournal(ctx context.Context, user entities.User) (entities.Journal, error)

	AddMark(ctx context.Context, mark entities.Mark) (entities.Lesson, error)
	GetMark(ctx context.Context, group string, subject string, userIdHex string, user entities.User) ([]entities.Lesson, error)
	UpdateMark(ctx context.Context, mark entities.Mark) (entities.Lesson, error)
	DeleteMark(ctx context.Context, markIdHex string, userIdHex string, subjectIdHex string) (entities.Lesson, error)
}

type journalController struct {
	repository repositories.JournalRepository
}

func NewJournalController(repository repositories.JournalRepository) JournalController {
	return &journalController{repository: repository}
}

func (j *journalController) GetJournalAvailableOptions(ctx context.Context, user entities.User) ([]entities.JournalAvailableOption, error) {
	if user.Type == "group" {
		return []entities.JournalAvailableOption{{
			Teacher:  "",
			Subject:  "",
			Group:    user.TypeName,
			Editable: false,
		}}, nil
	}

	options, err := j.repository.GetAvailableOptions(ctx, user.Name, utils.SliceContains(user.Permissions, "editJournal"))
	if err != nil {
		return nil, err
	}

	return options, nil
}

func (j *journalController) GetJournal(ctx context.Context, group string, subject string, teacher string, user entities.User) (entities.Journal, error) {
	if group == "" || subject == "" || teacher == "" {
		return entities.Journal{}, NotValidParams
	}

	return j.repository.GetJournal(ctx, group, subject, user.TypeName, user.StudyPlaceId)
}

func (j *journalController) GetUserJournal(ctx context.Context, user entities.User) (entities.Journal, error) {
	return j.repository.GetStudentJournal(ctx, user.Id, user.TypeName, user.StudyPlaceId)
}

func (j *journalController) AddMark(ctx context.Context, mark entities.Mark) (entities.Lesson, error) {
	if mark.Mark == "" || mark.UserId.IsZero() || mark.LessonId.IsZero() {
		return entities.Lesson{}, NotValidParams
	}

	if _, err := j.repository.AddMark(ctx, mark); err != nil {
		return entities.Lesson{}, err
	}

	return j.repository.GetLessonByIDAndUserID(ctx, mark.UserId, mark.LessonId)
}

func (j *journalController) GetMark(ctx context.Context, group string, subject string, userIdHex string, user entities.User) ([]entities.Lesson, error) {
	if group == "" || subject == "" || userIdHex == "" {
		return nil, NotValidParams
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if err != nil {
		return nil, err
	}

	return j.repository.GetLessons(ctx, userId, group, user.Name, subject, user.StudyPlaceId)
}

func (j *journalController) UpdateMark(ctx context.Context, mark entities.Mark) (entities.Lesson, error) {
	if mark.Mark == "" || mark.Id.IsZero() || mark.UserId.IsZero() || mark.LessonId.IsZero() {
		return entities.Lesson{}, NotValidParams
	}

	if err := j.repository.UpdateMark(ctx, mark); err != nil {
		return entities.Lesson{}, err
	}

	return j.repository.GetLessonByIDAndUserID(ctx, mark.UserId, mark.LessonId)
}

func (j *journalController) DeleteMark(ctx context.Context, markIdHex string, userIdHex string, subjectIdHex string) (entities.Lesson, error) {
	if markIdHex == "" || userIdHex == "" || subjectIdHex == "" {
		return entities.Lesson{}, NotValidParams
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if err != nil {
		return entities.Lesson{}, errors.Wrap(NotValidParams, "userId")
	}

	markId, err := primitive.ObjectIDFromHex(markIdHex)
	if err != nil {
		return entities.Lesson{}, errors.Wrap(NotValidParams, "markId")
	}

	subjectId, err := primitive.ObjectIDFromHex(subjectIdHex)
	if err != nil {
		return entities.Lesson{}, errors.Wrap(NotValidParams, "subjectId")
	}

	if err = j.repository.DeleteMarkByIDAndLessonID(ctx, markId, subjectId); err != nil {
		return entities.Lesson{}, err
	}

	return j.repository.GetLessonByIDAndUserID(ctx, userId, subjectId)
}
