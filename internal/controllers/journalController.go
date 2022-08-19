package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/dto"
	"studyum/internal/entities"
	parser "studyum/internal/parser/handler"
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

	AddMark(ctx context.Context, dto dto.AddMarkDTO, user entities.User) (entities.Mark, error)
	GetMark(ctx context.Context, group string, subject string, userIdHex string, user entities.User) ([]entities.Lesson, error)
	UpdateMark(ctx context.Context, dto dto.UpdateMarkDTO) error
	DeleteMark(ctx context.Context, markIdHex string, subjectIdHex string) error
}

type journalController struct {
	parser parser.Handler

	repository repositories.JournalRepository
}

func NewJournalController(parser parser.Handler, repository repositories.JournalRepository) JournalController {
	return &journalController{parser: parser, repository: repository}
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

func (j *journalController) AddMark(ctx context.Context, dto dto.AddMarkDTO, user entities.User) (entities.Mark, error) {
	if dto.Mark == "" || dto.UserId.IsZero() || dto.LessonId.IsZero() {
		return entities.Mark{}, NotValidParams
	}

	mark := entities.Mark{
		Mark:         dto.Mark,
		UserId:       dto.UserId,
		LessonId:     dto.LessonId,
		StudyPlaceId: user.StudyPlaceId,
	}

	id, err := j.repository.AddMark(ctx, mark)
	if err != nil {
		return entities.Mark{}, err
	}

	mark.Id = id
	mark.ParsedInfo = j.parser.AddMark(mark)
	if mark.ParsedInfo != nil {
		_ = j.repository.UpdateMark(ctx, mark)
	}

	return mark, nil
}

func (j *journalController) UpdateMark(ctx context.Context, dto dto.UpdateMarkDTO) error {
	if dto.Mark == "" || dto.Id.IsZero() || dto.LessonId.IsZero() {
		return NotValidParams
	}

	mark := entities.Mark{
		Id:       dto.Id,
		Mark:     dto.Mark,
		UserId:   dto.UserId,
		LessonId: dto.LessonId,
	}

	mark.ParsedInfo = j.parser.EditMark(mark)
	if err := j.repository.UpdateMark(ctx, mark); err != nil {
		return err
	}

	return nil
}

func (j *journalController) DeleteMark(ctx context.Context, markIdHex string, subjectIdHex string) error {
	if markIdHex == "" || subjectIdHex == "" {
		return NotValidParams
	}

	markId, err := primitive.ObjectIDFromHex(markIdHex)
	if err != nil {
		return errors.Wrap(NotValidParams, "markId")
	}

	subjectId, err := primitive.ObjectIDFromHex(subjectIdHex)
	if err != nil {
		return errors.Wrap(NotValidParams, "subjectId")
	}

	mark, err := j.repository.GetMarkById(ctx, markId)
	if err != nil {
		return err
	}

	if err = j.repository.DeleteMarkByIDAndLessonID(ctx, markId, subjectId); err != nil {
		return err
	}

	j.parser.DeleteMark(mark)
	return nil
}
