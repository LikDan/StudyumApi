package controllers

import (
	"context"
	"studyum/src/entities"
)

type IJournalController interface {
	GetJournalAvailableOptions(ctx context.Context, user entities.User) ([]entities.JournalAvailableOption, error)

	GetJournal(ctx context.Context, group string, subject string, teacher string, user entities.User) (entities.Journal, error)
	GetUserJournal(ctx context.Context, user entities.User) (entities.Journal, error)

	AddMark(ctx context.Context, mark entities.Mark, user entities.User) (entities.Lesson, error)
	GetMark(ctx context.Context, group string, subject string, userIdHex string, user entities.User) ([]entities.Lesson, error)
	UpdateMark(ctx context.Context, mark entities.Mark, user entities.User) (entities.Lesson, error)
	DeleteMark(ctx context.Context, markIdHex string, userIdHex string, subjectIdHex string, user entities.User) (entities.Lesson, error)
}
