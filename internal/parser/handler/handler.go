package handler

import (
	"context"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"studyum/internal/entities"
	"studyum/internal/parser/application"
	"studyum/internal/parser/controller"
	"studyum/internal/parser/dto"
	"studyum/pkg/firebase"
)

type Handler interface {
	Update(app application.App)

	AddMark(mark entities.Mark) map[string]any
	EditMark(mark entities.Mark) map[string]any
	DeleteMark(mark entities.Mark)

	AddLesson(lesson entities.Lesson) map[string]any
	EditLesson(lesson entities.Lesson) map[string]any
	DeleteLesson(lesson entities.Lesson)
}

type handler struct {
	firebase firebase.Firebase

	controller controller.Controller
}

func NewParserHandler(firebase firebase.Firebase, controller controller.Controller) Handler {
	h := &handler{firebase: firebase, controller: controller}

	for _, app := range controller.Apps() {
		ctx := context.Background()
		lastLesson, err := h.controller.GetLastLesson(ctx, app.StudyPlaceId())
		if err != nil {
			continue
		}

		app.Init(lastLesson)

		types := app.ScheduleTypesUpdate()
		if err = h.controller.InsertScheduleTypes(ctx, types); err != nil {
			continue
		}

		updateCron := cron.New()
		if err = updateCron.AddFunc(app.GetUpdateCronPattern(), func() { h.Update(app) }); err != nil {
			logrus.Warningf("cannot launch cron for %s, err: %e", app.GetName(), err)
			continue
		}

		updateCron.Start()
	}

	return h
}

func (h *handler) Update(app application.App) {
	ctx := context.Background()
	h.controller.Update(ctx, app)
}

func (h *handler) AddMark(mark entities.Mark) map[string]any {
	ctx := context.Background()

	markDTO := dto.Mark(mark)
	return h.controller.AddMark(ctx, markDTO)
}

func (h *handler) EditMark(mark entities.Mark) map[string]any {
	ctx := context.Background()

	markDTO := dto.Mark(mark)
	return h.controller.EditMark(ctx, markDTO)
}

func (h *handler) DeleteMark(mark entities.Mark) {
	ctx := context.Background()

	markDTO := dto.Mark(mark)
	h.controller.DeleteMark(ctx, markDTO)
}

func (h *handler) AddLesson(lesson entities.Lesson) map[string]any {
	ctx := context.Background()

	lessonDTO := dto.Lesson{
		Id:           lesson.Id,
		StudyPlaceId: lesson.StudyPlaceId,
		Type:         lesson.Type,
		EndDate:      lesson.EndDate,
		StartDate:    lesson.StartDate,
		Subject:      lesson.Subject,
		Group:        lesson.Group,
		Teacher:      lesson.Type,
		Room:         lesson.Room,
	}
	return h.controller.AddLesson(ctx, lessonDTO)
}

func (h *handler) EditLesson(lesson entities.Lesson) map[string]any {
	ctx := context.Background()

	lessonDTO := dto.Lesson{
		Id:           lesson.Id,
		StudyPlaceId: lesson.StudyPlaceId,
		Type:         lesson.Type,
		EndDate:      lesson.EndDate,
		StartDate:    lesson.StartDate,
		Subject:      lesson.Subject,
		Group:        lesson.Group,
		Teacher:      lesson.Type,
		Room:         lesson.Room,
	}
	return h.controller.EditLesson(ctx, lessonDTO)
}

func (h *handler) DeleteLesson(lesson entities.Lesson) {
	ctx := context.Background()

	lessonDTO := dto.Lesson{
		Id:           lesson.Id,
		StudyPlaceId: lesson.StudyPlaceId,
		Type:         lesson.Type,
		EndDate:      lesson.EndDate,
		StartDate:    lesson.StartDate,
		Subject:      lesson.Subject,
		Group:        lesson.Group,
		Teacher:      lesson.Type,
		Room:         lesson.Room,
	}
	h.controller.DeleteLesson(ctx, lessonDTO)
}
