package controllers

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
	"studyum/src/repositories"
	"studyum/src/utils"
)

type JournalController struct {
	repository repositories.IJournalRepository
}

func NewJournalController(repository repositories.IJournalRepository) *JournalController {
	return &JournalController{repository: repository}
}

func (j *JournalController) GetJournalAvailableOptions(ctx context.Context, user models.User) ([]models.JournalAvailableOption, *models.Error) {
	if user.Type == "group" {
		return []models.JournalAvailableOption{{
			Teacher:  "",
			Subject:  "",
			Group:    user.TypeName,
			Editable: false,
		}}, models.EmptyError()
	}

	options, err := j.repository.GetAvailableOptions(ctx, user.Name, utils.SliceContains(user.Permissions, "editJournal"))
	if err.Check() {
		return nil, err
	}

	return options, models.EmptyError()
}

func (j *JournalController) GetJournal(ctx context.Context, group string, subject string, teacher string, user models.User) (models.Journal, *models.Error) {
	if !utils.CheckNotEmpty(group, subject, teacher) {
		return models.Journal{}, models.BindErrorStr("provide valid params", 400, models.UNDEFINED)
	}

	var journal models.Journal
	if err := j.repository.GetJournal(ctx, &journal, group, subject, user.TypeName, user.StudyPlaceId); err.Check() {
		return models.Journal{}, err
	}

	return journal, models.EmptyError()
}

func (j *JournalController) GetUserJournal(ctx context.Context, user models.User) (models.Journal, *models.Error) {
	var journal models.Journal
	if err := j.repository.GetStudentJournal(ctx, &journal, user.Id, user.TypeName, user.StudyPlaceId); err.Check() {
		return models.Journal{}, err
	}

	return journal, models.EmptyError()
}

func (j *JournalController) AddMark(ctx context.Context, mark models.Mark, user models.User) (models.Lesson, *models.Error) {
	if !utils.SliceContains(user.Permissions, "editJournal") {
		return models.Lesson{}, models.BindErrorStr("no permission", 400, models.UNDEFINED)
	}

	if mark.Mark == "" || mark.UserId.IsZero() || mark.LessonId.IsZero() {
		return models.Lesson{}, models.BindErrorStr("provide valid params", 400, models.UNDEFINED)
	}

	if err := j.repository.AddMark(ctx, &mark); err.Check() {
		return models.Lesson{}, err
	}

	return j.repository.GetLessonById(ctx, mark.UserId, mark.LessonId)
}

func (j *JournalController) GetMark(ctx context.Context, group string, subject string, userIdHex string, user models.User) ([]models.Lesson, *models.Error) {
	teacher := user.Name

	if group == "" || subject == "" || userIdHex == "" {
		return nil, models.BindErrorStr("provide valid params", 400, models.UNDEFINED)
	}

	userId, err_ := primitive.ObjectIDFromHex(userIdHex)
	if err := models.BindError(err_, 400, models.UNDEFINED); err.Check() {
		return nil, err
	}

	return j.repository.GetLessons(ctx, userId, group, teacher, subject, user.StudyPlaceId)
}

func (j *JournalController) UpdateMark(ctx context.Context, mark models.Mark, user models.User) (models.Lesson, *models.Error) {
	if !utils.SliceContains(user.Permissions, "editJournal") {
		return models.Lesson{}, models.BindErrorStr("no permission", 400, models.UNDEFINED)
	}

	if mark.Mark == "" || mark.Id.IsZero() || mark.UserId.IsZero() || mark.LessonId.IsZero() {
		return models.Lesson{}, models.BindErrorStr("provide valid params", 400, models.UNDEFINED)
	}

	if err := j.repository.UpdateMark(ctx, &mark); err.Check() {
		return models.Lesson{}, err
	}

	return j.repository.GetLessonById(ctx, mark.UserId, mark.LessonId)
}

func (j *JournalController) DeleteMark(ctx context.Context, markIdHex string, userIdHex string, subjectIdHex string, user models.User) (models.Lesson, *models.Error) {
	if !utils.SliceContains(user.Permissions, "editJournal") {
		return models.Lesson{}, models.BindErrorStr("no permission", 400, models.UNDEFINED)
	}

	if markIdHex == "" || userIdHex == "" || subjectIdHex == "" {
		return models.Lesson{}, models.BindErrorStr("provide valid params", 400, models.UNDEFINED)
	}

	userId, err_ := primitive.ObjectIDFromHex(userIdHex)
	if err := models.BindError(err_, 400, models.UNDEFINED); err.Check() {
		return models.Lesson{}, err
	}

	markId, err_ := primitive.ObjectIDFromHex(markIdHex)
	if err := models.BindError(err_, 400, models.UNDEFINED); err.Check() {
		return models.Lesson{}, err
	}

	subjectId, err_ := primitive.ObjectIDFromHex(subjectIdHex)
	if err := models.BindError(err_, 400, models.UNDEFINED); err.Check() {
		return models.Lesson{}, err
	}

	if err := j.repository.DeleteMark(ctx, markId, subjectId); err.Check() {
		return models.Lesson{}, err
	}

	return j.repository.GetLessonById(ctx, userId, subjectId)
}
