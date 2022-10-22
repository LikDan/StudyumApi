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

	AddAbsence(ctx context.Context, absencesDTO dto.AddAbsencesDTO, user entities.User) (entities.Absence, error)
	UpdateAbsence(ctx context.Context, user entities.User, absences dto.UpdateAbsencesDTO) error
	DeleteAbsence(ctx context.Context, user entities.User, id string) error
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

	journal, err := j.repository.GetJournal(ctx, group, subject, user.TypeName, user.StudyPlaceID)
	if err != nil {
		return entities.Journal{}, err
	}

	for i := range journal.Rows {
		journal.Rows[i].Title = j.encrypt.DecryptString(journal.Rows[i].Title)
	}

	slices.SortFunc(journal.Rows, func(el1, el2 entities.JournalRow) bool {
		return el1.Title < el2.Title
	})

	return journal, nil
}

func (j *journalController) GetUserJournal(ctx context.Context, user entities.User) (entities.Journal, error) {
	return j.repository.GetStudentJournal(ctx, user.Id, user.TypeName, user.StudyPlaceID)
}

func (j *journalController) GetMark(ctx context.Context, group string, subject string, userIdHex string, user entities.User) ([]entities.Lesson, error) {
	if group == "" || subject == "" || userIdHex == "" {
		return nil, NotValidParams
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if err != nil {
		return nil, err
	}

	return j.repository.GetLessons(ctx, userId, group, user.Name, subject, user.StudyPlaceID)
}

func (j *journalController) AddMark(ctx context.Context, dto dto.AddMarkDTO, user entities.User) (entities.Mark, error) {
	if dto.Mark == "" || dto.StudentID.IsZero() || dto.LessonId.IsZero() {
		return entities.Mark{}, NotValidParams
	}

	mark := entities.Mark{
		Mark:         dto.Mark,
		StudentID:    dto.StudentID,
		LessonID:     dto.LessonId,
		StudyPlaceID: user.StudyPlaceID,
	}

	id, err := j.repository.AddMark(ctx, mark, user.TypeName)
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
		LessonID:  dto.LessonId,
	}

	if err := j.repository.UpdateMark(ctx, mark, user.TypeName); err != nil {
		return err
	}

	go j.parser.EditMark(mark)

	return nil
}

func (j *journalController) DeleteMark(ctx context.Context, user entities.User, markIdHex string) error {
	markId, err := primitive.ObjectIDFromHex(markIdHex)
	if err != nil {
		return errors.Wrap(NotValidParams, "markId")
	}

	if err = j.repository.DeleteMarkByID(ctx, markId, user.TypeName); err != nil {
		return err
	}

	go j.parser.DeleteMark(markId, user.StudyPlaceID)
	return nil
}

func (j *journalController) AddAbsence(ctx context.Context, dto dto.AddAbsencesDTO, user entities.User) (entities.Absence, error) {
	if dto.StudentID.IsZero() || dto.LessonID.IsZero() {
		return entities.Absence{}, NotValidParams
	}

	absences := entities.Absence{
		Time:         dto.Time,
		StudentID:    dto.StudentID,
		LessonID:     dto.LessonID,
		StudyPlaceID: user.StudyPlaceID,
	}

	id, err := j.repository.AddAbsence(ctx, absences, user.TypeName)
	if err != nil {
		return entities.Absence{}, err
	}

	absences.Id = id

	return absences, nil
}

func (j *journalController) UpdateAbsence(ctx context.Context, user entities.User, dto dto.UpdateAbsencesDTO) error {
	if dto.Id.IsZero() || dto.LessonID.IsZero() {
		return NotValidParams
	}

	absences := entities.Absence{
		Id:           dto.Id,
		Time:         dto.Time,
		StudentID:    dto.StudentID,
		LessonID:     dto.LessonID,
		StudyPlaceID: user.StudyPlaceID,
	}

	if err := j.repository.UpdateAbsence(ctx, absences, user.TypeName); err != nil {
		return err
	}

	return nil
}

func (j *journalController) DeleteAbsence(ctx context.Context, user entities.User, idHex string) error {
	absenceID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return errors.Wrap(NotValidParams, "markId")
	}

	if err = j.repository.DeleteAbsenceByID(ctx, absenceID, user.TypeName); err != nil {
		return err
	}

	return nil
}
