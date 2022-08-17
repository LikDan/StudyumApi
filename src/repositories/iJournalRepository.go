package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
)

type IJournalRepository interface {
	AddMark(ctx context.Context, mark *models.Mark) error
	UpdateMark(ctx context.Context, mark *models.Mark) error
	DeleteMark(ctx context.Context, id primitive.ObjectID, lessonId primitive.ObjectID) error

	GetAvailableOptions(ctx context.Context, teacher string, editable bool) ([]models.JournalAvailableOption, error)

	GetStudentJournal(ctx context.Context, journal *models.Journal, userId primitive.ObjectID, group string, studyPlaceId int) error
	GetJournal(ctx context.Context, journal *models.Journal, group string, subject string, typeName string, studyPlaceId int) error

	GetLessonById(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (models.Lesson, error)
	GetLessons(ctx context.Context, userId primitive.ObjectID, group, teacher, subject string, studyPlaceId int) ([]models.Lesson, error)
}
