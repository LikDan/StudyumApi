package controllers

import (
	"context"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"strconv"
	"studyum/internal/global"
	"studyum/internal/journal/dtos"
	"studyum/internal/journal/entities"
	"studyum/internal/journal/repositories"
	"studyum/internal/parser/dto"
	parser "studyum/internal/parser/handler"
	"studyum/internal/utils"
	"studyum/pkg/encryption"
)

type Controller interface {
	AddMarks(ctx context.Context, marks []dtos.AddMarkDTO, user global.User) ([]entities.Mark, error)
	AddMark(ctx context.Context, dto dtos.AddMarkDTO, user global.User) (entities.CellResponse, error)
	UpdateMark(ctx context.Context, user global.User, dto dtos.UpdateMarkDTO) (entities.CellResponse, error)
	DeleteMark(ctx context.Context, user global.User, markIdHex string) (entities.CellResponse, error)

	AddAbsences(ctx context.Context, dto []dtos.AddAbsencesDTO, user global.User) ([]entities.Absence, error)
	AddAbsence(ctx context.Context, absencesDTO dtos.AddAbsencesDTO, user global.User) (entities.CellResponse, error)
	UpdateAbsence(ctx context.Context, user global.User, absences dtos.UpdateAbsencesDTO) (entities.CellResponse, error)
	DeleteAbsence(ctx context.Context, user global.User, id string) (entities.CellResponse, error)

	GenerateMarksReport(ctx context.Context, config dtos.MarksReport, user global.User) (*excelize.File, error)
	GenerateAbsencesReport(ctx context.Context, config dtos.AbsencesReport, user global.User) (*excelize.File, error)
}

type controller struct {
	journal Journal

	parser     parser.Handler
	repository repositories.Repository
	encrypt    encryption.Encryption
}

func NewController(parser parser.Handler, journal Journal, repository repositories.Repository, encrypt encryption.Encryption) Controller {
	return &controller{parser: parser, journal: journal, repository: repository, encrypt: encrypt}
}

func (j *controller) GenerateMarksReport(ctx context.Context, config dtos.MarksReport, user global.User) (*excelize.File, error) {
	table, err := j.repository.GenerateMarksReport(ctx, user.TuitionGroup, config.LessonType, config.Mark, config.StartDate, config.EndDate, user.StudyPlaceID)
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
	err = f.SetCellValue(sheetName, "B1", user.TypeName+" -> "+user.TuitionGroup)

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

func (j *controller) GenerateAbsencesReport(ctx context.Context, config dtos.AbsencesReport, user global.User) (*excelize.File, error) {
	table, err := j.repository.GenerateAbsencesReport(ctx, user.TuitionGroup, config.StartDate, config.EndDate, user.StudyPlaceID)
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
	err = f.SetCellValue(sheetName, "B1", user.TypeName+" -> "+user.TuitionGroup)

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

func (j *controller) checkMarkExistence(ctx context.Context, mark dtos.AddMarkDTO, studyPlaceID primitive.ObjectID) bool {
	lesson, err := j.repository.GetLessonByID(ctx, mark.LessonID)
	if err != nil {
		return false
	}
	studyPlace, err := j.repository.GetStudyPlaceByID(ctx, studyPlaceID)
	if err != nil {
		return false
	}

	for _, lessonType := range studyPlace.LessonTypes {
		if lessonType.Type != lesson.Type {
			continue
		}

		for _, markType := range lessonType.Marks {
			if markType.Mark == mark.Mark {
				return true
			}
		}
		for _, markType := range lessonType.StandaloneMarks {
			if markType.Mark == mark.Mark {
				return true
			}
		}
	}

	return false
}

func (j *controller) AddMarks(ctx context.Context, addDTO []dtos.AddMarkDTO, user global.User) ([]entities.Mark, error) {
	marks := make([]entities.Mark, len(addDTO))
	for i, markDTO := range addDTO {
		if markDTO.Mark == "" || markDTO.StudentID.IsZero() || markDTO.LessonID.IsZero() || !j.checkMarkExistence(ctx, markDTO, user.StudyPlaceID) {
			return nil, global.NotValidParams
		}

		mark := entities.Mark{
			ID:           primitive.NewObjectID(),
			Mark:         markDTO.Mark,
			StudentID:    markDTO.StudentID,
			LessonID:     markDTO.LessonID,
			StudyPlaceID: user.StudyPlaceID,
		}

		id, err := j.repository.AddMark(ctx, mark, user.TypeName)
		if err != nil {
			return nil, err
		}
		mark.ID = id

		parserDTO := dto.MarkDTO{
			Id:           mark.ID,
			Mark:         mark.Mark,
			StudentID:    mark.StudentID,
			LessonId:     mark.LessonID,
			StudyPlaceId: mark.StudyPlaceID,
			ParsedInfo:   mark.ParsedInfo,
		}
		go j.parser.AddMark(parserDTO)

		marks[i] = mark
	}

	return marks, nil
}

func (j *controller) AddMark(ctx context.Context, addDTO dtos.AddMarkDTO, user global.User) (entities.CellResponse, error) {
	if addDTO.Mark == "" || addDTO.StudentID.IsZero() || addDTO.LessonID.IsZero() || !j.checkMarkExistence(ctx, addDTO, user.StudyPlaceID) {
		return entities.CellResponse{}, global.NotValidParams
	}

	mark := entities.Mark{
		Mark:         addDTO.Mark,
		StudentID:    addDTO.StudentID,
		LessonID:     addDTO.LessonID,
		StudyPlaceID: user.StudyPlaceID,
	}

	id, err := j.repository.AddMark(ctx, mark, user.TypeName)
	if err != nil {
		return entities.CellResponse{}, err
	}

	mark.ID = id

	parserDTO := dto.MarkDTO{
		Id:           mark.ID,
		Mark:         mark.Mark,
		StudentID:    mark.StudentID,
		LessonId:     mark.LessonID,
		StudyPlaceId: mark.StudyPlaceID,
		ParsedInfo:   mark.ParsedInfo,
	}
	go j.parser.AddMark(parserDTO)

	return j.journal.GetUpdateInfo(ctx, mark.StudentID, mark.LessonID)
}

func (j *controller) UpdateMark(ctx context.Context, user global.User, updateDTO dtos.UpdateMarkDTO) (entities.CellResponse, error) {
	if updateDTO.Mark == "" || updateDTO.ID.IsZero() || updateDTO.LessonID.IsZero() || !j.checkMarkExistence(ctx, updateDTO.AddMarkDTO, user.StudyPlaceID) {
		return entities.CellResponse{}, global.NotValidParams
	}

	mark := entities.Mark{
		ID:        updateDTO.ID,
		Mark:      updateDTO.Mark,
		StudentID: updateDTO.StudentID,
		LessonID:  updateDTO.LessonID,
	}

	if err := j.repository.UpdateMark(ctx, mark, user.TypeName); err != nil {
		return entities.CellResponse{}, err
	}

	parserDTO := dto.MarkDTO{
		Id:           mark.ID,
		Mark:         mark.Mark,
		StudentID:    mark.StudentID,
		LessonId:     mark.LessonID,
		StudyPlaceId: mark.StudyPlaceID,
		ParsedInfo:   mark.ParsedInfo,
	}
	go j.parser.EditMark(parserDTO)

	return j.journal.GetUpdateInfo(ctx, mark.StudentID, mark.LessonID)
}

func (j *controller) DeleteMark(ctx context.Context, user global.User, markIdHex string) (entities.CellResponse, error) {
	markId, err := primitive.ObjectIDFromHex(markIdHex)
	if err != nil {
		return entities.CellResponse{}, errors.Wrap(global.NotValidParams, "markId")
	}

	mark, err := j.repository.GetMarkByID(ctx, markId)
	if err != nil {
		return entities.CellResponse{}, err
	}

	if err = j.repository.DeleteMarkByID(ctx, markId, user.TypeName); err != nil {
		return entities.CellResponse{}, err
	}

	go j.parser.DeleteMark(markId, user.StudyPlaceID)
	return j.journal.GetUpdateInfo(ctx, mark.StudentID, mark.LessonID)
}

func (j *controller) AddAbsences(ctx context.Context, dto []dtos.AddAbsencesDTO, user global.User) ([]entities.Absence, error) {
	absences := make([]entities.Absence, len(dto))
	for i, markDTO := range dto {
		if markDTO.StudentID.IsZero() || markDTO.LessonID.IsZero() {
			return nil, global.NotValidParams
		}

		absence := entities.Absence{
			ID:           primitive.NewObjectID(),
			Time:         markDTO.Time,
			StudentID:    markDTO.StudentID,
			LessonID:     markDTO.LessonID,
			StudyPlaceID: user.StudyPlaceID,
		}

		id, err := j.repository.AddAbsence(ctx, absence, user.TypeName)
		if err != nil {
			return nil, err
		}
		absence.ID = id
		//TODO notify parser

		absences[i] = absence
	}

	return absences, nil
}

func (j *controller) AddAbsence(ctx context.Context, dto dtos.AddAbsencesDTO, user global.User) (entities.CellResponse, error) {
	if dto.StudentID.IsZero() || dto.LessonID.IsZero() {
		return entities.CellResponse{}, global.NotValidParams
	}

	absence := entities.Absence{
		Time:         dto.Time,
		StudentID:    dto.StudentID,
		LessonID:     dto.LessonID,
		StudyPlaceID: user.StudyPlaceID,
	}

	_, err := j.repository.AddAbsence(ctx, absence, user.TypeName)
	if err != nil {
		return entities.CellResponse{}, err
	}

	return j.journal.GetUpdateInfo(ctx, absence.StudentID, absence.LessonID)
}

func (j *controller) UpdateAbsence(ctx context.Context, user global.User, dto dtos.UpdateAbsencesDTO) (entities.CellResponse, error) {
	if dto.ID.IsZero() || dto.LessonID.IsZero() {
		return entities.CellResponse{}, global.NotValidParams
	}

	absence := entities.Absence{
		ID:           dto.ID,
		Time:         dto.Time,
		StudentID:    dto.StudentID,
		LessonID:     dto.LessonID,
		StudyPlaceID: user.StudyPlaceID,
	}

	if err := j.repository.UpdateAbsence(ctx, absence, user.TypeName); err != nil {
		return entities.CellResponse{}, err
	}
	return j.journal.GetUpdateInfo(ctx, absence.StudentID, absence.LessonID)
}

func (j *controller) DeleteAbsence(ctx context.Context, user global.User, idHex string) (entities.CellResponse, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return entities.CellResponse{}, errors.Wrap(global.NotValidParams, "markId")
	}

	absence, err := j.repository.GetAbsenceByID(ctx, id)
	if err != nil {
		return entities.CellResponse{}, err
	}

	if err = j.repository.DeleteAbsenceByID(ctx, id, user.TypeName); err != nil {
		return entities.CellResponse{}, err
	}

	return j.journal.GetUpdateInfo(ctx, absence.StudentID, absence.LessonID)
}
