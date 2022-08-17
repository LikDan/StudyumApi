package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/entities"
)

type IJournalRepository interface {
	AddMark(ctx context.Context, mark *entities.Mark) error
	UpdateMark(ctx context.Context, mark *entities.Mark) error
	DeleteMark(ctx context.Context, id primitive.ObjectID, lessonId primitive.ObjectID) error

	GetAvailableOptions(ctx context.Context, teacher string, editable bool) ([]entities.JournalAvailableOption, error)

	GetStudentJournal(ctx context.Context, journal *entities.Journal, userId primitive.ObjectID, group string, studyPlaceId int) error
	GetJournal(ctx context.Context, journal *entities.Journal, group string, subject string, typeName string, studyPlaceId int) error

	GetLessonById(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (entities.Lesson, error)
	GetLessons(ctx context.Context, userId primitive.ObjectID, group, teacher, subject string, studyPlaceId int) ([]entities.Lesson, error)
}
