package controllers

import (
	"context"
	"studyum/src/models"
)

type IJournalController interface {
	GetJournalAvailableOptions(ctx context.Context, user models.User) ([]models.JournalAvailableOption, *models.Error)

	GetJournal(ctx context.Context, group string, subject string, teacher string, user models.User) (models.Journal, *models.Error)
	GetUserJournal(ctx context.Context, user models.User) (models.Journal, *models.Error)

	AddMark(ctx context.Context, mark models.Mark, user models.User) (models.Lesson, *models.Error)
	GetMark(ctx context.Context, group string, subject string, userIdHex string, user models.User) ([]models.Lesson, *models.Error)
	UpdateMark(ctx context.Context, mark models.Mark, user models.User) (models.Lesson, *models.Error)
	DeleteMark(ctx context.Context, markIdHex string, userIdHex string, subjectIdHex string, user models.User) (models.Lesson, *models.Error)
}
