package kbp

import (
	"context"
	"github.com/sirupsen/logrus"
	"studyum/internal/parser/appDTO"
	"studyum/internal/parser/entities"
)

func (a *app) OnMarkAdd(_ context.Context, mark entities.Mark, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO {
	logrus.Infof("set mark %v, with lesson %v", mark, lesson)
	return nil
}

func (a *app) OnMarkEdit(_ context.Context, mark entities.Mark, lesson entities.Lesson) appDTO.ParsedInfoTypeDTO {
	logrus.Infof("edit mark %v, with lesson %v", mark, lesson)
	return nil
}

func (a *app) OnMarkDelete(_ context.Context, mark entities.Mark, lesson entities.Lesson) {
	logrus.Infof("delete mark %v, with lesson %v", mark, lesson)
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
