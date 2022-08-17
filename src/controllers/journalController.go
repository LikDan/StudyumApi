package controllers

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
	"studyum/src/repositories"
	"studyum/src/utils"
)

var (
	NotValidParams = errors.New("not valid params")
	NoPermission   = errors.New("no permission")
)

type JournalController struct {
	repository repositories.IJournalRepository
}

func NewJournalController(repository repositories.IJournalRepository) *JournalController {
	return &JournalController{repository: repository}
}

func (j *JournalController) GetJournalAvailableOptions(ctx context.Context, user models.User) ([]models.JournalAvailableOption, error) {
	if user.Type == "group" {
		return []models.JournalAvailableOption{{
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

func (j *JournalController) GetJournal(ctx context.Context, group string, subject string, teacher string, user models.User) (models.Journal, error) {
	if !utils.CheckNotEmpty(group, subject, teacher) {
		return models.Journal{}, NotValidParams
	}

	var journal models.Journal
	if err := j.repository.GetJournal(ctx, &journal, group, subject, user.TypeName, user.StudyPlaceId); err != nil {
		return models.Journal{}, err
	}

	return journal, nil
}

func (j *JournalController) GetUserJournal(ctx context.Context, user models.User) (models.Journal, error) {
	var journal models.Journal
	if err := j.repository.GetStudentJournal(ctx, &journal, user.Id, user.TypeName, user.StudyPlaceId); err != nil {
		return models.Journal{}, err
	}

	return journal, nil
}

func (j *JournalController) AddMark(ctx context.Context, mark models.Mark, user models.User) (models.Lesson, error) {
	if !utils.SliceContains(user.Permissions, "editJournal") {
		return models.Lesson{}, NoPermission
	}

	if mark.Mark == "" || mark.UserId.IsZero() || mark.LessonId.IsZero() {
		return models.Lesson{}, NotValidParams
	}

	if err := j.repository.AddMark(ctx, &mark); err != nil {
		return models.Lesson{}, err
	}

	return j.repository.GetLessonById(ctx, mark.UserId, mark.LessonId)
}

func (j *JournalController) GetMark(ctx context.Context, group string, subject string, userIdHex string, user models.User) ([]models.Lesson, error) {
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

func (j *JournalController) UpdateMark(ctx context.Context, mark models.Mark, user models.User) (models.Lesson, error) {
	if !utils.SliceContains(user.Permissions, "editJournal") {
		return models.Lesson{}, NoPermission
	}

	if mark.Mark == "" || mark.Id.IsZero() || mark.UserId.IsZero() || mark.LessonId.IsZero() {
		return models.Lesson{}, NotValidParams
	}

	if err := j.repository.UpdateMark(ctx, &mark); err != nil {
		return models.Lesson{}, err
	}

	return j.repository.GetLessonById(ctx, mark.UserId, mark.LessonId)
}

func (j *JournalController) DeleteMark(ctx context.Context, markIdHex string, userIdHex string, subjectIdHex string, user models.User) (models.Lesson, error) {
	if !utils.SliceContains(user.Permissions, "editJournal") {
		return models.Lesson{}, NoPermission
	}

	if markIdHex == "" || userIdHex == "" || subjectIdHex == "" {
		return models.Lesson{}, NotValidParams
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if err != nil {
		return models.Lesson{}, err
	}

	markId, err := primitive.ObjectIDFromHex(markIdHex)
	if err != nil {
		return models.Lesson{}, err
	}

	subjectId, err := primitive.ObjectIDFromHex(subjectIdHex)
	if err != nil {
		return models.Lesson{}, err
	}

	if err := j.repository.DeleteMark(ctx, markId, subjectId); err != nil {
		return models.Lesson{}, err
	}

	return j.repository.GetLessonById(ctx, userId, subjectId)
}
