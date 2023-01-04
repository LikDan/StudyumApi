package journal

import (
	"context"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"strconv"
	"studyum/internal/global"
	"studyum/internal/parser/dto"
	parser "studyum/internal/parser/handler"
	"studyum/internal/utils"
	"studyum/pkg/encryption"
)

type Controller interface {
	GetJournalAvailableOptions(ctx context.Context, user global.User) ([]AvailableOption, error)

	GetJournal(ctx context.Context, group string, subject string, teacher string, user global.User) (Journal, error)
	GetUserJournal(ctx context.Context, user global.User) (Journal, error)

	AddMarks(ctx context.Context, marks []AddMarkDTO, user global.User) ([]Mark, error)
	AddMark(ctx context.Context, dto AddMarkDTO, user global.User) (Mark, error)
	GetMark(ctx context.Context, group string, subject string, userIdHex string, user global.User) ([]Lesson, error)
	UpdateMark(ctx context.Context, user global.User, dto UpdateMarkDTO) error
	DeleteMark(ctx context.Context, user global.User, markIdHex string) error

	AddAbsences(ctx context.Context, dto []AddAbsencesDTO, user global.User) ([]Absence, error)
	AddAbsence(ctx context.Context, absencesDTO AddAbsencesDTO, user global.User) (Absence, error)
	UpdateAbsence(ctx context.Context, user global.User, absences UpdateAbsencesDTO) error
	DeleteAbsence(ctx context.Context, user global.User, id string) error

	Generate(ctx context.Context, config MarksReport, user global.User) (*excelize.File, error)
	GenerateAbsences(ctx context.Context, config AbsencesReport, user global.User) (*excelize.File, error)
}

type controller struct {
	parser parser.Handler

	repository Repository

	encrypt encryption.Encryption
}

func NewJournalController(parser parser.Handler, repository Repository, encrypt encryption.Encryption) Controller {
	return &controller{parser: parser, repository: repository, encrypt: encrypt}
}

func (j *controller) Generate(ctx context.Context, config MarksReport, user global.User) (*excelize.File, error) {
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

func (j *controller) GenerateAbsences(ctx context.Context, config AbsencesReport, user global.User) (*excelize.File, error) {
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

func (j *controller) GetJournalAvailableOptions(ctx context.Context, user global.User) ([]AvailableOption, error) {
	if user.Type == "group" {
		return []AvailableOption{{
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

func (j *controller) GetJournal(ctx context.Context, group string, subject string, teacher string, user global.User) (Journal, error) {
	if group == "" || subject == "" || teacher == "" {
		return Journal{}, global.NotValidParams
	}

	options, err := j.GetJournalAvailableOptions(ctx, user)
	if err != nil {
		return Journal{}, err
	}

	var option *AvailableOption
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
		return Journal{}, global.NoPermission
	}

	journal, err := j.repository.GetJournal(ctx, *option, user.StudyPlaceID)
	if err != nil {
		return Journal{}, err
	}

	for i := range journal.Rows {
		journal.Rows[i].Title = j.encrypt.DecryptString(journal.Rows[i].Title)
	}

	slices.SortFunc(journal.Rows, func(el1, el2 Row) bool {
		return el1.Title < el2.Title
	})

	return journal, nil
}

func (j *controller) GetUserJournal(ctx context.Context, user global.User) (Journal, error) {
	return j.repository.GetStudentJournal(ctx, user.Id, user.TypeName, user.StudyPlaceID)
}

func (j *controller) checkMarkExistence(ctx context.Context, mark AddMarkDTO, studyPlaceID primitive.ObjectID) bool {
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

func (j *controller) AddMarks(ctx context.Context, addDTO []AddMarkDTO, user global.User) ([]Mark, error) {
	marks := make([]Mark, len(addDTO))
	for i, markDTO := range addDTO {
		if markDTO.Mark == "" || markDTO.StudentID.IsZero() || markDTO.LessonID.IsZero() || !j.checkMarkExistence(ctx, markDTO, user.StudyPlaceID) {
			return nil, global.NotValidParams
		}

		mark := Mark{
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

func (j *controller) GetMark(ctx context.Context, group string, subject string, userIdHex string, user global.User) ([]Lesson, error) {
	if group == "" || subject == "" || userIdHex == "" {
		return nil, global.NotValidParams
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if err != nil {
		return nil, err
	}

	return j.repository.GetLessons(ctx, userId, group, user.Name, subject, user.StudyPlaceID)
}

func (j *controller) AddMark(ctx context.Context, addDTO AddMarkDTO, user global.User) (Mark, error) {
	if addDTO.Mark == "" || addDTO.StudentID.IsZero() || addDTO.LessonID.IsZero() || !j.checkMarkExistence(ctx, addDTO, user.StudyPlaceID) {
		return Mark{}, global.NotValidParams
	}

	mark := Mark{
		Mark:         addDTO.Mark,
		StudentID:    addDTO.StudentID,
		LessonID:     addDTO.LessonID,
		StudyPlaceID: user.StudyPlaceID,
	}

	id, err := j.repository.AddMark(ctx, mark, user.TypeName)
	if err != nil {
		return Mark{}, err
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

	return mark, nil
}

func (j *controller) UpdateMark(ctx context.Context, user global.User, updateDTO UpdateMarkDTO) error {
	if updateDTO.Mark == "" || updateDTO.ID.IsZero() || updateDTO.LessonID.IsZero() || !j.checkMarkExistence(ctx, updateDTO.AddMarkDTO, user.StudyPlaceID) {
		return global.NotValidParams
	}

	mark := Mark{
		ID:        updateDTO.ID,
		Mark:      updateDTO.Mark,
		StudentID: updateDTO.StudentID,
		LessonID:  updateDTO.LessonID,
	}

	if err := j.repository.UpdateMark(ctx, mark, user.TypeName); err != nil {
		return err
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

	return nil
}

func (j *controller) DeleteMark(ctx context.Context, user global.User, markIdHex string) error {
	markId, err := primitive.ObjectIDFromHex(markIdHex)
	if err != nil {
		return errors.Wrap(global.NotValidParams, "markId")
	}

	if err = j.repository.DeleteMarkByID(ctx, markId, user.TypeName); err != nil {
		return err
	}

	go j.parser.DeleteMark(markId, user.StudyPlaceID)
	return nil
}

func (j *controller) AddAbsences(ctx context.Context, dto []AddAbsencesDTO, user global.User) ([]Absence, error) {
	absences := make([]Absence, len(dto))
	for i, markDTO := range dto {
		if markDTO.StudentID.IsZero() || markDTO.LessonID.IsZero() {
			return nil, global.NotValidParams
		}

		absence := Absence{
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

func (j *controller) AddAbsence(ctx context.Context, dto AddAbsencesDTO, user global.User) (Absence, error) {
	if dto.StudentID.IsZero() || dto.LessonID.IsZero() {
		return Absence{}, global.NotValidParams
	}

	absences := Absence{
		Time:         dto.Time,
		StudentID:    dto.StudentID,
		LessonID:     dto.LessonID,
		StudyPlaceID: user.StudyPlaceID,
	}

	id, err := j.repository.AddAbsence(ctx, absences, user.TypeName)
	if err != nil {
		return Absence{}, err
	}

	absences.ID = id

	return absences, nil
}

func (j *controller) UpdateAbsence(ctx context.Context, user global.User, dto UpdateAbsencesDTO) error {
	if dto.ID.IsZero() || dto.LessonID.IsZero() {
		return global.NotValidParams
	}

	absences := Absence{
		ID:           dto.ID,
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

func (j *controller) DeleteAbsence(ctx context.Context, user global.User, idHex string) error {
	absenceID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return errors.Wrap(global.NotValidParams, "markId")
	}

	if err = j.repository.DeleteAbsenceByID(ctx, absenceID, user.TypeName); err != nil {
		return err
	}

	return nil
}
