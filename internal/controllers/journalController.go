package controllers

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
	"studyum/internal/repositories"
	"studyum/internal/utils"
)

var (
	NotValidParams = errors.New("not valid params")
	NoPermission   = errors.New("no permission")
)

type JournalController struct {
	repository repositories.JournalRepository
}

func NewJournalController(repository repositories.JournalRepository) *JournalController {
	return &JournalController{repository: repository}
}

func (j *JournalController) GetJournalAvailableOptions(ctx context.Context, user entities.User) ([]entities.JournalAvailableOption, error) {
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

func (j *JournalController) GetJournal(ctx context.Context, group string, subject string, teacher string, user entities.User) (entities.Journal, error) {
	if group == "" || subject == "" || teacher == "" {
		return entities.Journal{}, NotValidParams
	}

	var journal entities.Journal
	if _, err := j.repository.GetJournal(ctx, group, subject, user.TypeName, user.StudyPlaceId); err != nil {
		return entities.Journal{}, err
	}

	return journal, nil
}

func (j *JournalController) GetUserJournal(ctx context.Context, user entities.User) (entities.Journal, error) {
	var journal entities.Journal
	if _, err := j.repository.GetStudentJournal(ctx, user.Id, user.TypeName, user.StudyPlaceId); err != nil {
		return entities.Journal{}, err
	}

	return journal, nil
}

func (j *JournalController) AddMark(ctx context.Context, mark entities.Mark, user entities.User) (entities.Lesson, error) {
	if !utils.SliceContains(user.Permissions, "editJournal") {
		return entities.Lesson{}, NoPermission
	}

	if mark.Mark == "" || mark.UserId.IsZero() || mark.LessonId.IsZero() {
		return entities.Lesson{}, NotValidParams
	}

	if _, err := j.repository.AddMark(ctx, mark); err != nil {
		return entities.Lesson{}, err
	}

	return j.repository.GetLessonByIDAndUserID(ctx, mark.UserId, mark.LessonId)
}

func (j *JournalController) GetMark(ctx context.Context, group string, subject string, userIdHex string, user entities.User) ([]entities.Lesson, error) {
	teacher := user.Name

	if group == "" || subject == "" || userIdHex == "" {
		return nil, NotValidParams
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if err != nil {
		return nil, err
	}

	return j.repository.GetLessons(ctx, userId, group, teacher, subject, user.StudyPlaceId)
}

func (j *JournalController) UpdateMark(ctx context.Context, mark entities.Mark, user entities.User) (entities.Lesson, error) {
	if !utils.SliceContains(user.Permissions, "editJournal") {
		return entities.Lesson{}, NoPermission
	}

	if mark.Mark == "" || mark.Id.IsZero() || mark.UserId.IsZero() || mark.LessonId.IsZero() {
		return entities.Lesson{}, NotValidParams
	}

	if err := j.repository.UpdateMark(ctx, mark); err != nil {
		return entities.Lesson{}, err
	}

	return j.repository.GetLessonByIDAndUserID(ctx, mark.UserId, mark.LessonId)
}

func (j *JournalController) DeleteMark(ctx context.Context, markIdHex string, userIdHex string, subjectIdHex string, user entities.User) (entities.Lesson, error) {
	if !utils.SliceContains(user.Permissions, "editJournal") {
		return entities.Lesson{}, NoPermission
	}

	if markIdHex == "" || userIdHex == "" || subjectIdHex == "" {
		return entities.Lesson{}, NotValidParams
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if err != nil {
		return entities.Lesson{}, err
	}

	markId, err := primitive.ObjectIDFromHex(markIdHex)
	if err != nil {
		return entities.Lesson{}, err
	}

	subjectId, err := primitive.ObjectIDFromHex(subjectIdHex)
	if err != nil {
		return entities.Lesson{}, err
	}

	if err := j.repository.DeleteMarkByIDAndLessonID(ctx, markId, subjectId); err != nil {
		return entities.Lesson{}, err
	}

	return j.repository.GetLessonByIDAndUserID(ctx, userId, subjectId)
}
