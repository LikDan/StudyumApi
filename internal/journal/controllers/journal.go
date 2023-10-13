package controllers

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"strconv"
	auth "studyum/internal/auth/entities"
	general "studyum/internal/general/entities"
	"studyum/internal/journal/entities"
	"studyum/internal/journal/repositories"
	"studyum/internal/utils"
	"studyum/pkg/encryption"
	"time"
)

type Journal interface {
	GetUpdateInfo(ctx context.Context, userID, lessonID primitive.ObjectID) (entities.CellResponse, error)
	GetAvailableOptions(ctx context.Context, user auth.User) ([]entities.CategoryOptions, error)

	BuildAvailableOptions(ctx context.Context, user auth.User) ([]entities.AvailableOption, error)
	BuildSubjectsJournal(ctx context.Context, group string, subject string, teacher string, user auth.User) (entities.Journal, error)
	BuildStudentsJournal(ctx context.Context, user auth.User) (entities.Journal, error)
}

type journal struct {
	repository repositories.Repository
	encrypt    encryption.Encryption
}

func NewJournalController(repository repositories.Repository, encrypt encryption.Encryption) Journal {
	return &journal{repository: repository, encrypt: encrypt}
}

func (c *journal) rowAverageMark(row entities.Row) float32 {
	sum := 0
	amount := 0
	for _, c := range row.Cells {
		if c == nil {
			continue
		}

		for _, m := range c.Marks {
			mark, _ := strconv.Atoi(m.Mark)
			if mark != 0 {
				sum += mark
				amount++
			}
		}
	}

	if amount == 0 {
		return 0
	}

	return float32(sum) / float32(amount)
}

func (c *journal) markColor(colorSet general.JournalColors, date time.Time, role general.LessonType, mark entities.Mark) string {
	var markType general.MarkType
	for _, markType_ := range role.Marks {
		if markType_.Mark == mark.Mark {
			markType = markType_
		}
	}

	if markType.Mark == "" {
		for _, markType_ := range role.StandaloneMarks {
			if markType_.Mark == mark.Mark {
				markType = markType_
			}
		}
	}

	if markType.Mark == "" {
		return colorSet.General
	}

	if markType.WorkOutTime == 0 {
		return colorSet.General
	}

	now := time.Now()
	if date.Add(markType.WorkOutTime).After(now) {
		return colorSet.Warning
	}

	return colorSet.Danger
}

func (c *journal) absenceColor(colorSet general.JournalColors, date time.Time, role general.LessonType, absence entities.Absence) string {
	if absence.Time != nil {
		return colorSet.General
	}

	now := time.Now()
	if date.Add(role.AbsenceWorkOutTime).After(now) {
		return colorSet.Warning
	}

	return colorSet.Danger
}

func (c *journal) cellColor(studyPlace general.StudyPlace, date time.Time, cell entities.Cell) string {
	cellType := cell.Type[0]
	var role = general.LessonType{}
	for _, t := range studyPlace.LessonTypes {
		if cellType == t.Type {
			role = t
		}
	}

	if role.Type == "" {
		return studyPlace.JournalColors.General
	}

	color := studyPlace.JournalColors.General
	for _, m := range cell.Marks {
		markColor := c.markColor(studyPlace.JournalColors, date, role, m)
		if markColor == studyPlace.JournalColors.General {
			return markColor
		}

		if markColor == studyPlace.JournalColors.Danger || (markColor == studyPlace.JournalColors.Warning && color == studyPlace.JournalColors.General) {
			color = markColor
		}
	}
	if color == studyPlace.JournalColors.Danger {
		return color
	}

	if len(cell.Absences) != 0 {
		if color := c.absenceColor(studyPlace.JournalColors, date, role, cell.Absences[0]); color != studyPlace.JournalColors.General {
			return color
		}
	}

	return color
}

func (c *journal) rowColor(studyPlace general.StudyPlace, row entities.Row) string {
	color := studyPlace.JournalColors.General
	for _, cell := range row.Cells {
		if cell == nil {
			continue
		}

		if cell.JournalCellColor == studyPlace.JournalColors.Danger {
			return cell.JournalCellColor
		}

		if cell.JournalCellColor == studyPlace.JournalColors.Warning {
			color = cell.JournalCellColor
		}
	}

	return color
}

func (c *journal) rowMarksAmount(cells []*entities.Cell) map[string]int {
	marks := map[string]int{}
	for _, cell := range cells {
		if cell == nil {
			continue
		}

		for _, mark := range cell.Marks {
			marks[mark.Mark]++
		}
	}

	return marks
}

func (c *journal) rowAbsencesAmount(cells []*entities.Cell) (int, int) {
	absences := 0
	lateness := 0
	for _, cell := range cells {
		if cell == nil {
			continue
		}

		for _, absence := range cell.Absences {
			if absence.Time == nil {
				absences++
			} else {
				lateness += *absence.Time
			}
		}
	}

	return absences, lateness
}

func (c *journal) proceedJournal(j *entities.Journal) {
	for i := range j.Rows {
		for ci, cell := range j.Rows[i].Cells {
			if cell == nil {
				continue
			}

			cell.JournalCellColor = c.cellColor(j.Info.StudyPlace, j.Dates[ci].StartDate, *cell)
		}
		j.Rows[i].AverageMark = c.rowAverageMark(j.Rows[i])
		j.Rows[i].Color = c.rowColor(j.Info.StudyPlace, j.Rows[i])
		j.Rows[i].MarksAmount = c.rowMarksAmount(j.Rows[i].Cells)
		j.Rows[i].AbsencesAmount, j.Rows[i].AbsencesTime = c.rowAbsencesAmount(j.Rows[i].Cells)
	}
}

func (c *journal) getCellByLessonAndUserID(ctx context.Context, userID primitive.ObjectID, lesson entities.Lesson) (entities.Cell, error) {
	cell := entities.Cell{
		Id:   lesson.Id,
		Type: []string{lesson.Type},
	}

	for _, mark := range lesson.Marks {
		if mark.StudentID == userID {
			cell.Marks = append(cell.Marks, mark)
		}
	}

	for _, absence := range lesson.Absences {
		if absence.StudentID == userID {
			cell.Absences = append(cell.Absences, absence)
		}
	}

	studyPlace, err := c.repository.GetStudyPlaceByID(ctx, lesson.StudyPlaceId)
	if err != nil {
		return entities.Cell{}, err
	}

	cell.JournalCellColor = c.cellColor(studyPlace, lesson.StartDate, cell)
	return cell, nil
}

func (c *journal) getRowInfo(ctx context.Context, userID primitive.ObjectID, lesson entities.Lesson) (float32, map[string]int, string, error) {
	studyPlace, err := c.repository.GetStudyPlaceByID(ctx, lesson.StudyPlaceId)
	if err != nil {
		return 0, nil, "", err
	}

	cells, dates, err := c.repository.GetJournalRowWithDates(ctx, userID, lesson.Subject, lesson.Teacher, lesson.Group, lesson.StudyPlaceId)
	if err != nil {
		return 0, nil, "", err
	}

	row := entities.Row{
		Cells: make([]*entities.Cell, len(cells)),
	}

	for i, cell := range cells {
		if cell == nil {
			continue
		}

		cell.JournalCellColor = c.cellColor(studyPlace, dates[i], *cell)
		row.Cells[i] = cell
	}

	return c.rowAverageMark(row), c.rowMarksAmount(row.Cells), c.rowColor(studyPlace, row), nil
}

func (c *journal) GetUpdateInfo(ctx context.Context, userID, lessonID primitive.ObjectID) (entities.CellResponse, error) {
	lesson, err := c.repository.GetLessonByID(ctx, lessonID)
	if err != nil {
		return entities.CellResponse{}, err
	}

	cell, err := c.getCellByLessonAndUserID(ctx, userID, lesson)
	if err != nil {
		return entities.CellResponse{}, err
	}

	avg, amount, color, err := c.getRowInfo(ctx, userID, lesson)
	if err != nil {
		return entities.CellResponse{}, err
	}

	return entities.CellResponse{
		Cell:       cell,
		RowColor:   color,
		Average:    avg,
		MarkAmount: amount,
	}, nil
}

func (c *journal) GetAvailableOptions(ctx context.Context, user auth.User) ([]entities.CategoryOptions, error) {
	options, err := c.BuildAvailableOptions(ctx, user)
	if err != nil {
		return nil, err
	}

	return []entities.CategoryOptions{{Category: "selectOption", Options: options}}, nil
}

func (c *journal) BuildAvailableOptions(ctx context.Context, user auth.User) ([]entities.AvailableOption, error) {
	if user.StudyPlaceInfo.Role == "student" {
		return []entities.AvailableOption{{
			Header:     "myJournal",
			Teacher:    "",
			Subject:    "",
			Group:      "",
			Editable:   false,
			HasMembers: true,
		}}, nil
	}

	var options []entities.AvailableOption
	appendOptions := func(opts []entities.AvailableOption) {
		for _, opt := range opts {
			found := false
			for i, option := range options {
				if option.Group == opt.Group && option.Subject == opt.Subject && option.Teacher == opt.Teacher {
					options[i].Editable = option.Editable || opt.Editable
					found = true
				}
			}

			if !found {
				options = append(options, opt)
			}
		}
	}

	teacherOptions, err := c.repository.GetAvailableOptions(ctx, user.StudyPlaceInfo.ID, user.StudyPlaceInfo.RoleName, slices.Contains(user.StudyPlaceInfo.Permissions, "editJournal"))
	if err != nil {
		return nil, err
	}

	appendOptions(teacherOptions)

	if tuitionOptions, err := c.repository.GetAvailableTuitionOptions(ctx, user.StudyPlaceInfo.ID, user.StudyPlaceInfo.TuitionGroup, false); err == nil {
		appendOptions(tuitionOptions)
	}

	if utils.HasPermission(user, "viewJournals") {
		adminOptions, err := c.repository.GetAllAvailableOptions(ctx, user.StudyPlaceInfo.ID, false)
		if err == nil {
			appendOptions(adminOptions)
		}
	}

	return options, nil
}

func (c *journal) BuildSubjectsJournal(ctx context.Context, group string, subject string, teacher string, user auth.User) (entities.Journal, error) {
	if group == "" || subject == "" || teacher == "" {
		return entities.Journal{}, NotValidParams
	}

	options, err := c.BuildAvailableOptions(ctx, user)
	if err != nil {
		return entities.Journal{}, err
	}

	var option *entities.AvailableOption
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
		return entities.Journal{}, ErrNoPermission
	}

	journal, err := c.repository.GetJournal(ctx, *option, user.StudyPlaceInfo.ID)
	if err != nil {
		return entities.Journal{}, err
	}

	for i := range journal.Rows {
		journal.Rows[i].Title = c.encrypt.DecryptString(journal.Rows[i].Title)
	}

	slices.SortFunc(journal.Rows, func(el1, el2 entities.Row) bool {
		return el1.Title < el2.Title
	})

	if err != nil {
		return entities.Journal{}, err
	}

	c.proceedJournal(&journal)
	return journal, nil
}

func (c *journal) BuildStudentsJournal(ctx context.Context, user auth.User) (entities.Journal, error) {
	journal, err := c.repository.GetStudentJournal(ctx, user.Id, user.StudyPlaceInfo.RoleName, user.StudyPlaceInfo.ID)
	if err != nil {
		return entities.Journal{}, err
	}

	c.proceedJournal(&journal)
	return journal, nil
}
