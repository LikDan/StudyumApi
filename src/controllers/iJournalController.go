package controllers

import (
	"context"
	"studyum/src/models"
)

type IJournalController interface {
	GetJournalAvailableOptions(ctx context.Context, user models.User) ([]models.JournalAvailableOption, error)

	GetJournal(ctx context.Context, group string, subject string, teacher string, user models.User) (models.Journal, error)
	GetUserJournal(ctx context.Context, user models.User) (models.Journal, error)

	AddMark(ctx context.Context, mark models.Mark, user models.User) (models.Lesson, error)
	GetMark(ctx context.Context, group string, subject string, userIdHex string, user models.User) ([]models.Lesson, error)
	UpdateMark(ctx context.Context, mark models.Mark, user models.User) (models.Lesson, error)
	DeleteMark(ctx context.Context, markIdHex string, userIdHex string, subjectIdHex string, user models.User) (models.Lesson, error)
}
