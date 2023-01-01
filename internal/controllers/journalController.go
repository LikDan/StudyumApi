package controllers

import (
	"context"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"strconv"
	"studyum/internal/dto"
	"studyum/internal/entities"
	parser "studyum/internal/parser/handler"
	"studyum/internal/repositories"
	"studyum/internal/utils"
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

	AddMarks(ctx context.Context, marks []dto.AddMarkDTO, user entities.User) ([]entities.Mark, error)
	AddMark(ctx context.Context, dto dto.AddMarkDTO, user entities.User) (entities.Mark, error)
	GetMark(ctx context.Context, group string, subject string, userIdHex string, user entities.User) ([]entities.Lesson, error)
	UpdateMark(ctx context.Context, user entities.User, dto dto.UpdateMarkDTO) error
	DeleteMark(ctx context.Context, user entities.User, markIdHex string) error

	AddAbsences(ctx context.Context, dto []dto.AddAbsencesDTO, user entities.User) ([]entities.Absence, error)
	AddAbsence(ctx context.Context, absencesDTO dto.AddAbsencesDTO, user entities.User) (entities.Absence, error)
	UpdateAbsence(ctx context.Context, user entities.User, absences dto.UpdateAbsencesDTO) error
	DeleteAbsence(ctx context.Context, user entities.User, id string) error

	Generate(ctx context.Context, config dto.MarksReport, user entities.User) (*excelize.File, error)
	GenerateAbsences(ctx context.Context, config dto.AbsencesReport, user entities.User) (*excelize.File, error)
}

type journalController struct {
	parser parser.Handler

	repository repositories.JournalRepository

	encrypt encryption.Encryption
}

func NewJournalController(parser parser.Handler, repository repositories.JournalRepository, encrypt encryption.Encryption) JournalController {
	return &journalController{parser: parser, repository: repository, encrypt: encrypt}
}

func (j *journalController) Generate(ctx context.Context, config dto.MarksReport, user entities.User) (*excelize.File, error) {
	table, err := j.repository.Generate(ctx, user.TuitionGroup, config.LessonType, config.Mark, config.StartDate, config.EndDate, user.StudyPlaceID)
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

func (j *journalController) GenerateAbsences(ctx context.Context, config dto.AbsencesReport, user entities.User) (*excelize.File, error) {
	table, err := j.repository.GenerateAbsences(ctx, user.TuitionGroup, config.StartDate, config.EndDate, user.StudyPlaceID)
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

	if tuitionOptions, err := j.repository.GetAvailableTuitionOptions(ctx, user.TuitionGroup, false); err == nil {
		options = append(options, tuitionOptions...)
	}

	return options, nil
}

func (j *journalController) GetJournal(ctx context.Context, group string, subject string, teacher string, user entities.User) (entities.Journal, error) {
	if group == "" || subject == "" || teacher == "" {
		return entities.Journal{}, NotValidParams
	}

	options, err := j.GetJournalAvailableOptions(ctx, user)
	if err != nil {
		return entities.Journal{}, err
	}

	var option *entities.JournalAvailableOption
	for _, opt := range options {
		if opt.Group == group && opt.Teacher == teacher && opt.Subject == subject {
			var temp = opt
			option = &temp
		}
		if opt.Group == group && opt.Teacher == teacher && opt.Subject == subject && opt.Editable {
			var temp = opt
			option = &temp
			break
		}
	}

	if option == nil {
		return entities.Journal{}, NoPermission
	}

	journal, err := j.repository.GetJournal(ctx, *option, user.StudyPlaceID)
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

func (j *journalController) checkMarkExistence(ctx context.Context, mark dto.AddMarkDTO, studyPlaceID primitive.ObjectID) bool {
	lesson, err := j.repository.GetLessonByID(ctx, mark.LessonId)
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

func (j *journalController) AddMarks(ctx context.Context, dto []dto.AddMarkDTO, user entities.User) ([]entities.Mark, error) {
	marks := make([]entities.Mark, len(dto))
	for i, markDTO := range dto {
		if markDTO.Mark == "" || markDTO.StudentID.IsZero() || markDTO.LessonId.IsZero() || !j.checkMarkExistence(ctx, markDTO, user.StudyPlaceID) {
			return nil, NotValidParams
		}

		mark := entities.Mark{
			Id:           primitive.NewObjectID(),
			Mark:         markDTO.Mark,
			StudentID:    markDTO.StudentID,
			LessonID:     markDTO.LessonId,
			StudyPlaceID: user.StudyPlaceID,
		}

		id, err := j.repository.AddMark(ctx, mark, user.TypeName)
		if err != nil {
			return nil, err
		}
		mark.Id = id
		go j.parser.AddMark(mark)

		marks[i] = mark
	}

	return marks, nil
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
	if dto.Mark == "" || dto.StudentID.IsZero() || dto.LessonId.IsZero() || !j.checkMarkExistence(ctx, dto, user.StudyPlaceID) {
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
	if dto.Mark == "" || dto.Id.IsZero() || dto.LessonId.IsZero() || !j.checkMarkExistence(ctx, dto.AddMarkDTO, user.StudyPlaceID) {
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

func (j *journalController) AddAbsences(ctx context.Context, dto []dto.AddAbsencesDTO, user entities.User) ([]entities.Absence, error) {
	absences := make([]entities.Absence, len(dto))
	for i, markDTO := range dto {
		if markDTO.StudentID.IsZero() || markDTO.LessonID.IsZero() {
			return nil, NotValidParams
		}

		absence := entities.Absence{
			Id:           primitive.NewObjectID(),
			Time:         markDTO.Time,
			StudentID:    markDTO.StudentID,
			LessonID:     markDTO.LessonID,
			StudyPlaceID: user.StudyPlaceID,
		}

		id, err := j.repository.AddAbsence(ctx, absence, user.TypeName)
		if err != nil {
			return nil, err
		}
		absence.Id = id
		//TODO notify parser

		absences[i] = absence
	}

	return absences, nil
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
