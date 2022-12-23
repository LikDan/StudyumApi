package kbp

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/appDTO"
	"studyum/internal/parser/entities"
)

func (a *app) OnMarkAdd(ctx context.Context, mark entities.Mark, lesson entities.Lesson, student entities.User) appDTO.ParsedInfoTypeDTO {
	a.controller.AddMark(ctx, mark, lesson, student)
	logrus.Infof("set mark %v, with lesson %v", mark, lesson)
	return nil
}

func (a *app) OnMarkEdit(_ context.Context, mark entities.Mark, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO {
	logrus.Infof("edit mark %v, with lesson %v", mark, lesson)
	return nil
}

func (a *app) OnMarkDelete(_ context.Context, id primitive.ObjectID) {
	logrus.Infof("delete mark with id %v", id)
}

func (a *app) OnLessonAdd(_ context.Context, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO {
	logrus.Infof("add lesson %v", lesson)
	return nil
}

func (a *app) OnLessonEdit(_ context.Context, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO {
	logrus.Infof("add edit %v", lesson)
	return nil
}

func (a *app) OnLessonDelete(_ context.Context, lesson entities.Lesson) {
	logrus.Infof("add delete %v", lesson)
}
