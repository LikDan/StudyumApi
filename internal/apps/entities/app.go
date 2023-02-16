package entities

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/apps/shared"
	journalEntities "studyum/internal/journal/entities"
	scheduleEntities "studyum/internal/schedule/entities"
)

type LessonsManageInterface interface {
	AddLesson(ctx context.Context, data shared.Data, lesson scheduleEntities.Lesson) shared.Data
	UpdateLesson(ctx context.Context, data shared.Data, lesson scheduleEntities.Lesson) shared.Data
	RemoveLesson(ctx context.Context, data shared.Data, lesson scheduleEntities.Lesson)
}

type MarksManageInterface interface {
	AddMark(ctx context.Context, data shared.Data, mark journalEntities.Mark) shared.Data
	UpdateMark(ctx context.Context, data shared.Data, mark journalEntities.Mark) shared.Data
	RemoveMark(ctx context.Context, data shared.Data, mark journalEntities.Mark)
}

type AbsencesManageInterface interface {
	AddAbsence(ctx context.Context, data shared.Data, absence journalEntities.Absence) shared.Data
	UpdateAbsence(ctx context.Context, data shared.Data, absence journalEntities.Absence) shared.Data
	RemoveAbsence(ctx context.Context, data shared.Data, absence journalEntities.Absence)
}

type App interface {
	Init(shared shared.Shared)
	GetStudyPlaceID(ctx context.Context) primitive.ObjectID
}
