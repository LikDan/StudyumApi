package controllers

import (
	"context"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"strconv"
	apps "studyum/internal/apps/controllers"
	auth "studyum/internal/auth/entities"
	"studyum/internal/journal/dtos"
	"studyum/internal/journal/entities"
	"studyum/internal/journal/repositories"
	"studyum/internal/utils"
	"studyum/pkg/encryption"
)

var NotValidParams = errors.New("not valid params")
var ErrNoPermission = errors.New("no permission")

type Controller interface {
	AddMarks(ctx context.Context, marks []dtos.AddMarkDTO, user auth.User) ([]entities.StudentMark, error)
	AddMark(ctx context.Context, dto dtos.AddMarkDTO, user auth.User) (entities.JournalLesson, error)
	UpdateMark(ctx context.Context, user auth.User, dto dtos.UpdateMarkDTO) (entities.JournalLesson, error)
	DeleteMark(ctx context.Context, user auth.User, markIdHex string) (entities.JournalLesson, error)

	AddAbsences(ctx context.Context, dto []dtos.AddAbsencesDTO, user auth.User) ([]entities.Absence, error)
	AddAbsence(ctx context.Context, absencesDTO dtos.AddAbsencesDTO, user auth.User) (entities.JournalLesson, error)
	UpdateAbsence(ctx context.Context, user auth.User, absences dtos.UpdateAbsencesDTO) (entities.JournalLesson, error)
	DeleteAbsence(ctx context.Context, user auth.User, id string) (entities.JournalLesson, error)

	GenerateMarksReport(ctx context.Context, config dtos.MarksReport, user auth.User) (*excelize.File, error)
	GenerateAbsencesReport(ctx context.Context, config dtos.AbsencesReport, user auth.User) (*excelize.File, error)

	GetLessonInfo(ctx context.Context, user auth.User, studentID, id string) (entities.JournalLesson, error)
}

type controller struct {
	journal Journal

	apps       apps.Controller
	repository repositories.Repository
	encrypt    encryption.Encryption
}

func NewController(journal Journal, repository repositories.Repository, encrypt encryption.Encryption, apps apps.Controller) Controller {
	return &controller{journal: journal, apps: apps, repository: repository, encrypt: encrypt}
}

func (j *controller) GenerateMarksReport(ctx context.Context, config dtos.MarksReport, user auth.User) (*excelize.File, error) {
	table, err := j.repository.GenerateMarksReport(ctx, user.StudyPlaceInfo.TuitionGroup, config.LessonType, config.Mark, config.StartDate, config.EndDate, user.StudyPlaceInfo.ID)
	if err != nil {
		return nil, err
	}

	for i := range table.Rows {
		table.Rows[i][0] = j.encrypt.DecryptString(table.Rows[i][0])
	}

	slices.SortFunc(table.Rows, func(el1, el2 []string) bool {
		return el1[0] < el2[0]
	})

	f := excelize.NewFile()
	sheetName := f.GetSheetList()[0]

	if err = f.MergeCell(sheetName, "B1", "D1"); err != nil {
		return nil, err
	}
	err = f.SetCellValue(sheetName, "B1", user.StudyPlaceInfo.RoleName+" -> "+user.StudyPlaceInfo.TuitionGroup)

	column := "B"
	for _, title := range table.Titles {
		if err = f.SetCellValue(sheetName, column+"3", title); err != nil {
			return nil, err
		}
		column = utils.NextColumn(column)
	}

	for y, row := range table.Rows {
		column = "B"
		for _, el := range row {
			if err = f.SetCellValue(sheetName, column+strconv.Itoa(y+4), el); err != nil {
				return nil, err
			}
			column = utils.NextColumn(column)
		}
	}

	if err = utils.AutoSizeColumns(f, sheetName); err != nil {
		return nil, err
	}

	return f, nil
}

func (j *controller) GenerateAbsencesReport(ctx context.Context, config dtos.AbsencesReport, user auth.User) (*excelize.File, error) {
	table, err := j.repository.GenerateAbsencesReport(ctx, user.StudyPlaceInfo.TuitionGroup, config.StartDate, config.EndDate, user.StudyPlaceInfo.ID)
	if err != nil {
		return nil, err
	}

	for i := range table.Rows {
		table.Rows[i][0] = j.encrypt.DecryptString(table.Rows[i][0])
	}

	slices.SortFunc(table.Rows, func(el1, el2 []string) bool {
		return el1[0] < el2[0]
	})

	f := excelize.NewFile()
	sheetName := f.GetSheetList()[0]

	if err = f.MergeCell(sheetName, "B1", "D1"); err != nil {
		return nil, err
	}
	err = f.SetCellValue(sheetName, "B1", user.StudyPlaceInfo.RoleName+" -> "+user.StudyPlaceInfo.TuitionGroup)

	column := "B"
	for _, title := range table.Titles {
		if err = f.SetCellValue(sheetName, column+"3", title); err != nil {
			return nil, err
		}
		column = utils.NextColumn(column)
	}

	for y, row := range table.Rows {
		column = "B"
		for _, el := range row {
			if err = f.SetCellValue(sheetName, column+strconv.Itoa(y+4), el); err != nil {
				return nil, err
			}
			column = utils.NextColumn(column)
		}
	}

	if err = utils.AutoSizeColumns(f, sheetName); err != nil {
		return nil, err
	}

	return f, nil
}

func (j *controller) AddMarks(ctx context.Context, addDTO []dtos.AddMarkDTO, user auth.User) ([]entities.StudentMark, error) {
	marks := make([]entities.StudentMark, len(addDTO))
	for i, markDTO := range addDTO {
		if markDTO.MarkID.IsZero() || markDTO.StudentID.IsZero() || markDTO.LessonID.IsZero() {
			return nil, NotValidParams
		}

		mark := entities.StudentMark{
			ID:           primitive.NewObjectID(),
			MarkID:       markDTO.MarkID,
			StudentID:    markDTO.StudentID,
			LessonID:     markDTO.LessonID,
			StudyPlaceID: user.StudyPlaceInfo.ID,
		}

		if err := j.repository.AddMark(ctx, mark); err != nil {
			return nil, err
		}

		j.apps.AsyncEvent(user.StudyPlaceInfo.ID, "AddMark", mark)

		marks[i] = mark
	}

	return marks, nil
}

func (j *controller) AddMark(ctx context.Context, addDTO dtos.AddMarkDTO, user auth.User) (entities.JournalLesson, error) {
	if addDTO.MarkID.IsZero() || addDTO.StudentID.IsZero() || addDTO.LessonID.IsZero() {
		return entities.JournalLesson{}, NotValidParams
	}

	mark := entities.StudentMark{
		ID:           primitive.NewObjectID(),
		MarkID:       addDTO.MarkID,
		StudentID:    addDTO.StudentID,
		LessonID:     addDTO.LessonID,
		StudyPlaceID: user.StudyPlaceInfo.ID,
	}

	if err := j.repository.AddMark(ctx, mark); err != nil {
		return entities.JournalLesson{}, err
	}

	j.apps.AsyncEvent(user.StudyPlaceInfo.ID, "AddMark", mark)

	return j.repository.GetStudentLessonByID(ctx, mark.StudentID, mark.LessonID)
}

func (j *controller) UpdateMark(ctx context.Context, user auth.User, updateDTO dtos.UpdateMarkDTO) (entities.JournalLesson, error) {
	if updateDTO.MarkID.IsZero() || updateDTO.ID.IsZero() || updateDTO.LessonID.IsZero() {
		return entities.JournalLesson{}, NotValidParams
	}

	mark := entities.StudentMark{
		ID:        updateDTO.ID,
		MarkID:    updateDTO.MarkID,
		StudentID: updateDTO.StudentID,
		LessonID:  updateDTO.LessonID,
	}

	if err := j.repository.UpdateMark(ctx, mark); err != nil {
		return entities.JournalLesson{}, err
	}

	j.apps.AsyncEvent(user.StudyPlaceInfo.ID, "UpdateMark", mark)

	return j.repository.GetStudentLessonByID(ctx, mark.StudentID, mark.LessonID)
}

func (j *controller) DeleteMark(ctx context.Context, user auth.User, markIdHex string) (entities.JournalLesson, error) {
	markID, err := primitive.ObjectIDFromHex(markIdHex)
	if err != nil || markID.IsZero() {
		return entities.JournalLesson{}, err
	}

	mark, err := j.repository.GetMarkByID(ctx, markID)
	if err != nil {
		return entities.JournalLesson{}, err
	}

	j.apps.Event(user.StudyPlaceInfo.ID, "RemoveMark", mark)

	err = j.repository.DeleteMarkByID(ctx, markID)
	if err != nil {
		return entities.JournalLesson{}, err
	}

	return j.repository.GetStudentLessonByID(ctx, mark.StudentID, mark.LessonID)
}

func (j *controller) AddAbsences(ctx context.Context, dto []dtos.AddAbsencesDTO, user auth.User) ([]entities.Absence, error) {
	absences := make([]entities.Absence, len(dto))
	for i, markDTO := range dto {
		if markDTO.StudentID.IsZero() || markDTO.LessonID.IsZero() {
			return nil, NotValidParams
		}

		absence := entities.Absence{
			ID:           primitive.NewObjectID(),
			Time:         markDTO.Time,
			StudentID:    markDTO.StudentID,
			LessonID:     markDTO.LessonID,
			StudyPlaceID: user.StudyPlaceInfo.ID,
		}

		if err := j.repository.AddAbsence(ctx, absence); err != nil {
			return nil, err
		}

		j.apps.AsyncEvent(user.StudyPlaceInfo.ID, "AddAbsence", absence)

		absences[i] = absence
	}

	return absences, nil
}

func (j *controller) AddAbsence(ctx context.Context, dto dtos.AddAbsencesDTO, user auth.User) (entities.JournalLesson, error) {
	if dto.StudentID.IsZero() || dto.LessonID.IsZero() {
		return entities.JournalLesson{}, NotValidParams
	}

	absence := entities.Absence{
		ID:           primitive.NewObjectID(),
		Time:         dto.Time,
		StudentID:    dto.StudentID,
		LessonID:     dto.LessonID,
		StudyPlaceID: user.StudyPlaceInfo.ID,
	}

	err := j.repository.AddAbsence(ctx, absence)
	if err != nil {
		return entities.JournalLesson{}, err
	}

	j.apps.AsyncEvent(user.StudyPlaceInfo.ID, "AddAbsence", absence)

	return j.repository.GetStudentLessonByID(ctx, absence.StudentID, absence.LessonID)
}

func (j *controller) UpdateAbsence(ctx context.Context, user auth.User, dto dtos.UpdateAbsencesDTO) (entities.JournalLesson, error) {
	if dto.ID.IsZero() || dto.LessonID.IsZero() {
		return entities.JournalLesson{}, NotValidParams
	}

	absence := entities.Absence{
		ID:           dto.ID,
		Time:         dto.Time,
		StudentID:    dto.StudentID,
		LessonID:     dto.LessonID,
		StudyPlaceID: user.StudyPlaceInfo.ID,
	}

	if err := j.repository.UpdateAbsence(ctx, absence); err != nil {
		return entities.JournalLesson{}, err
	}

	j.apps.AsyncEvent(user.StudyPlaceInfo.ID, "UpdateAbsence", absence)

	return j.repository.GetStudentLessonByID(ctx, absence.StudentID, absence.LessonID)
}

func (j *controller) DeleteAbsence(ctx context.Context, user auth.User, idHex string) (entities.JournalLesson, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return entities.JournalLesson{}, err
	}

	absence, err := j.repository.GetAbsenceByID(ctx, id)
	if err != nil {
		return entities.JournalLesson{}, err
	}

	j.apps.AsyncEvent(user.StudyPlaceInfo.ID, "RemoveAbsence", entities.DeleteAbsenceID{ID: absence.ID})

	err = j.repository.DeleteAbsenceByID(ctx, id)
	if err != nil {
		return entities.JournalLesson{}, err
	}

	return j.repository.GetStudentLessonByID(ctx, absence.StudentID, absence.LessonID)
}

func (j *controller) GetLessonInfo(ctx context.Context, user auth.User, studentIDHex, idHex string) (entities.JournalLesson, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return entities.JournalLesson{}, errors.Wrap(NotValidParams, "markId")
	}

	studentID, err := primitive.ObjectIDFromHex(studentIDHex)
	if err != nil {
		return entities.JournalLesson{}, errors.Wrap(NotValidParams, "markId")
	}

	//todo check access to student ID
	return j.repository.GetStudentLessonByID(ctx, studentID, id)
}
