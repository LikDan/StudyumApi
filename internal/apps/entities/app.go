package entities

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	journalEntities "studyum/internal/journal/entities"
	scheduleEntities "studyum/internal/schedule/entities"
)

type Data map[string]any

type LessonsManageInterface interface {
	AddLesson(ctx context.Context, data Data, lesson scheduleEntities.Lesson) Data
	UpdateLesson(ctx context.Context, data Data, lesson scheduleEntities.Lesson) Data
	RemoveLesson(ctx context.Context, data Data, id primitive.ObjectID)
}

type MarksManageInterface interface {
	AddMark(ctx context.Context, data Data, mark journalEntities.Mark) Data
	UpdateMark(ctx context.Context, data Data, mark journalEntities.Mark) Data
	RemoveMark(ctx context.Context, data Data, id primitive.ObjectID)
}

type AbsencesManageInterface interface {
	AddAbsence(ctx context.Context, data Data, absence journalEntities.Absence) Data
	UpdateAbsence(ctx context.Context, data Data, absence journalEntities.Absence) Data
	RemoveAbsence(ctx context.Context, data Data, id primitive.ObjectID)
}

type App interface {
	GetStudyPlaceID(ctx context.Context) primitive.ObjectID
}
