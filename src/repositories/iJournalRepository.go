package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
)

type IJournalRepository interface {
	AddMark(ctx context.Context, mark *models.Mark) *models.Error
	UpdateMark(ctx context.Context, mark *models.Mark) *models.Error
	DeleteMark(ctx context.Context, id primitive.ObjectID, lessonId primitive.ObjectID) *models.Error

	GetAvailableOptions(ctx context.Context, teacher string, editable bool) ([]models.JournalAvailableOption, *models.Error)

	GetStudentJournal(ctx context.Context, journal *models.Journal, userId primitive.ObjectID, group string, studyPlaceId int) *models.Error
	GetJournal(ctx context.Context, journal *models.Journal, group string, subject string, typeName string, studyPlaceId int) *models.Error

	GetLessonById(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (models.Lesson, *models.Error)
	GetLessons(ctx context.Context, userId primitive.ObjectID, group, teacher, subject string, studyPlaceId int) ([]models.Lesson, *models.Error)
}
