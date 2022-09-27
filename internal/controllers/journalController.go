package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"studyum/internal/dto"
	"studyum/internal/entities"
	parser "studyum/internal/parser/handler"
	"studyum/internal/repositories"
	"studyum/pkg/encryption"
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
	UpdateMark(ctx context.Context, user entities.User, dto dto.UpdateMarkDTO) error
	DeleteMark(ctx context.Context, user entities.User, markIdHex string) error
}

type journalController struct {
	parser parser.Handler

	repository repositories.JournalRepository

	encrypt encryption.Encryption
}

func NewJournalController(parser parser.Handler, repository repositories.JournalRepository, encrypt encryption.Encryption) JournalController {
	return &journalController{parser: parser, repository: repository, encrypt: encrypt}
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

	options, err := j.repository.GetAvailableOptions(ctx, user.TypeName, slices.Contains(user.Permissions, "editJournal"))
	if err != nil {
		return nil, err
	}

	return options, nil
}

func (j *journalController) GetJournal(ctx context.Context, group string, subject string, teacher string, user entities.User) (entities.Journal, error) {
	if group == "" || subject == "" || teacher == "" {
		return entities.Journal{}, NotValidParams
	}

	journal, err := j.repository.GetJournal(ctx, group, subject, user.TypeName, user.StudyPlaceId)
	if err != nil {
		return entities.Journal{}, err
	}

	j.encrypt.Decrypt(&journal)
	return journal, nil
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
	if dto.Mark == "" || dto.StudentID.IsZero() || dto.LessonId.IsZero() {
		return entities.Mark{}, NotValidParams
	}

	mark := entities.Mark{
		Mark:         dto.Mark,
		StudentID:    dto.StudentID,
		LessonId:     dto.LessonId,
		StudyPlaceId: user.StudyPlaceId,
	}

	lesson, err := j.repository.GetLessonByID(ctx, mark.LessonId)
	if err != nil {
		return entities.Mark{}, err
	}

	if lesson.Teacher != user.TypeName {
		return entities.Mark{}, NoPermission
	}

	id, err := j.repository.AddMark(ctx, mark)
	if err != nil {
		return entities.Mark{}, err
	}

	mark.Id = id

	go j.parser.AddMark(mark)

	return mark, nil
}

func (j *journalController) UpdateMark(ctx context.Context, user entities.User, dto dto.UpdateMarkDTO) error {
	if dto.Mark == "" || dto.Id.IsZero() || dto.LessonId.IsZero() {
		return NotValidParams
	}

	mark := entities.Mark{
		Id:        dto.Id,
		Mark:      dto.Mark,
		StudentID: dto.StudentID,
		LessonId:  dto.LessonId,
	}

	lesson, err := j.repository.GetLessonByID(ctx, mark.LessonId)
	if err != nil {
		return err
	}

	if lesson.Teacher != user.TypeName {
		return NoPermission
	}

	if err = j.repository.UpdateMark(ctx, mark); err != nil {
		return err
	}

	go j.parser.EditMark(mark)

	return nil
}

func (j *journalController) DeleteMark(ctx context.Context, user entities.User, markIdHex string) error {
	if markIdHex == "" {
		return NotValidParams
	}

	markId, err := primitive.ObjectIDFromHex(markIdHex)
	if err != nil {
		return errors.Wrap(NotValidParams, "markId")
	}

	mark, err := j.repository.GetMarkById(ctx, markId)
	if err != nil {
		return err
	}

	lesson, err := j.repository.GetLessonByID(ctx, mark.LessonId)
	if err != nil {
		return err
	}

	if lesson.Teacher != user.TypeName {
		return NoPermission
	}

	if err = j.repository.DeleteMarkByID(ctx, markId); err != nil {
		return err
	}

	go j.parser.DeleteMark(mark)
	return nil
}
