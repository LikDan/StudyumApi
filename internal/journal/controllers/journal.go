package controllers

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"sort"
	auth "studyum/internal/auth/entities"
	general "studyum/internal/general/entities"
	"studyum/internal/journal/entities"
	"studyum/internal/journal/repositories"
	"studyum/internal/utils"
	"studyum/pkg/encryption"
	"time"
)

type Journal interface {
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

func (c *journal) decryptJournalRowTitle(j *entities.Journal) {
	for i := range j.RowTitles {
		j.RowTitles[i].Title = c.encrypt.DecryptString(j.RowTitles[i].Title)
	}
}

func (c *journal) sortJournal(j *entities.Journal) {
	sort.Slice(j.RowTitles, func(i1, i2 int) bool {
		return j.RowTitles[i1].Title <= j.RowTitles[i2].Title
	})

	sort.Slice(j.Dates, func(i1, i2 int) bool {
		return j.Dates[i1].Date.Unix() <= j.Dates[i2].Date.Unix()
	})
}

func (c *journal) assignJournalPoints(j *entities.Journal) {
	for i, cell := range j.Cells {
		x := slices.IndexFunc(j.Dates, func(date entities.JournalDate) bool {
			if date.ID.IsZero() {
				return date.Date == cell.Date.Date
			}
			return date.ID.Hex() == cell.Date.ID.Hex()
		})

		y := slices.IndexFunc(j.RowTitles, func(title entities.JournalRowTitle) bool {
			return title.ID.Hex() == cell.RowTitle.ID.Hex()
		})

		j.Cells[i].Point = entities.Point{
			X: x,
			Y: y,
		}
	}
}

func (c *journal) proceedJournal(j *entities.Journal) {
	c.sortJournal(j)
	c.assignJournalPoints(j)
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

	teacherOptions, err := c.repository.GetAvailableOptions(ctx, user.StudyPlaceInfo.ID, user.StudyPlaceInfo.TypeID, slices.Contains(user.StudyPlaceInfo.Permissions, "editJournal"))
	if err != nil {
		return nil, err
	}

	appendOptions(teacherOptions)

	groupID, err := c.repository.GetTypeID(ctx, user.StudyPlaceInfo.ID, "group", user.StudyPlaceInfo.RoleName)
	if err != nil {
		return options, nil
	}

	if tuitionOptions, err := c.repository.GetAvailableTuitionOptions(ctx, user.StudyPlaceInfo.ID, groupID, false); err == nil {
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

	groupID, err := primitive.ObjectIDFromHex(group)
	if err != nil {
		return entities.Journal{}, err
	}

	subjectID, err := primitive.ObjectIDFromHex(subject)
	if err != nil {
		return entities.Journal{}, err
	}

	teacherID, err := primitive.ObjectIDFromHex(teacher)
	if err != nil {
		return entities.Journal{}, err
	}

	//todo check permission
	_ = teacherID
	if false {
		return entities.Journal{}, ErrNoPermission
	}

	j, err := c.repository.GetJournal(ctx, user.StudyPlaceInfo.ID, groupID, subjectID)
	if err != nil {
		return entities.Journal{}, err
	}

	c.decryptJournalRowTitle(&j)
	c.proceedJournal(&j)

	//todo check permission
	j.Info.Editable = true

	return j, nil
}

func (c *journal) BuildStudentsJournal(ctx context.Context, user auth.User) (entities.Journal, error) {
	groupID, err := c.repository.GetTypeID(ctx, user.StudyPlaceInfo.ID, "group", user.StudyPlaceInfo.RoleName)
	if err != nil {
		return entities.Journal{}, err
	}

	j, err := c.repository.GetStudentJournal(ctx, user.Id, groupID, user.StudyPlaceInfo.ID)
	if err != nil {
		return entities.Journal{}, err
	}

	c.proceedJournal(&j)

	j.Info.Editable = false

	return j, nil
}
