package controller

import (
	"context"
	"studyum/src/parser/entities"
)

type IController interface {
	Apps() []entities.IApp

	UpdateGeneralSchedule(app entities.IApp)
	Update(ctx context.Context, app entities.IApp)
	GetLastLesson(ctx context.Context, id int) (entities.Lesson, error)
	InsertScheduleTypes(ctx context.Context, types []*entities.ScheduleTypeInfo) error
	GetAppByStudyPlaceId(id int) (entities.IApp, error)
}
