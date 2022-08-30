package handler

import (
	"context"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"studyum/internal/entities"
	"studyum/internal/parser/apps"
	"studyum/internal/parser/controller"
	"studyum/internal/parser/dto"
	"studyum/pkg/firebase"
)

type Handler interface {
	Update(app apps.App)

	AddMark(mark entities.Mark)
	EditMark(mark entities.Mark)
	DeleteMark(mark entities.Mark)

	AddLesson(lesson entities.Lesson)
	EditLesson(lesson entities.Lesson)
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

		types := app.ScheduleTypesUpdate()
		if err := h.controller.InsertScheduleTypes(ctx, types); err != nil {
			continue
		}

		updateCron := cron.New()
		if err := updateCron.AddFunc(app.GetUpdateCronPattern(), func() { h.Update(app) }); err != nil {
			logrus.Warningf("cannot launch cron for %s, err: %e", app.GetName(), err)
			continue
		}

		updateCron.Start()
	}

	return h
}

func (h *handler) Update(app apps.App) {
	ctx := context.Background()
	h.controller.Update(ctx, app)
}

func (h *handler) AddMark(mark entities.Mark) {
	ctx := context.Background()

	markDTO := dto.Mark{
		Id:           mark.Id,
		Mark:         mark.Mark,
		StudentID:    mark.StudentID,
		LessonId:     mark.LessonId,
		StudyPlaceId: mark.StudyPlaceId,
	}
	h.controller.AddMark(ctx, markDTO)
}

func (h *handler) EditMark(mark entities.Mark) {
	ctx := context.Background()

	markDTO := dto.Mark{
		Id:           mark.Id,
		Mark:         mark.Mark,
		StudentID:    mark.StudentID,
		LessonId:     mark.LessonId,
		StudyPlaceId: mark.StudyPlaceId,
	}
	h.controller.EditMark(ctx, markDTO)
}

func (h *handler) DeleteMark(mark entities.Mark) {
	ctx := context.Background()

	markDTO := dto.Mark{
		Id:           mark.Id,
		Mark:         mark.Mark,
		StudentID:    mark.StudentID,
		LessonId:     mark.LessonId,
		StudyPlaceId: mark.StudyPlaceId,
	}
	h.controller.DeleteMark(ctx, markDTO)
}

func (h *handler) AddLesson(lesson entities.Lesson) {
	ctx := context.Background()

	lessonDTO := dto.Lesson{
		Id:             lesson.Id,
		StudyPlaceId:   lesson.StudyPlaceId,
		PrimaryColor:   lesson.PrimaryColor,
		SecondaryColor: lesson.SecondaryColor,
		EndDate:        lesson.EndDate,
		StartDate:      lesson.StartDate,
		Subject:        lesson.Subject,
		Group:          lesson.Group,
		Teacher:        lesson.Teacher,
		Room:           lesson.Room,
	}
	h.controller.AddLesson(ctx, lessonDTO)
}

func (h *handler) EditLesson(lesson entities.Lesson) {
	ctx := context.Background()

	lessonDTO := dto.Lesson{
		Id:             lesson.Id,
		StudyPlaceId:   lesson.StudyPlaceId,
		PrimaryColor:   lesson.PrimaryColor,
		SecondaryColor: lesson.SecondaryColor,
		EndDate:        lesson.EndDate,
		StartDate:      lesson.StartDate,
		Subject:        lesson.Subject,
		Group:          lesson.Group,
		Teacher:        lesson.Teacher,
		Room:           lesson.Room,
	}
	h.controller.EditLesson(ctx, lessonDTO)
}

func (h *handler) DeleteLesson(lesson entities.Lesson) {
	ctx := context.Background()

	lessonDTO := dto.Lesson{
		Id:             lesson.Id,
		StudyPlaceId:   lesson.StudyPlaceId,
		PrimaryColor:   lesson.PrimaryColor,
		SecondaryColor: lesson.SecondaryColor,
		EndDate:        lesson.EndDate,
		StartDate:      lesson.StartDate,
		Subject:        lesson.Subject,
		Group:          lesson.Group,
		Teacher:        lesson.Teacher,
		Room:           lesson.Room,
	}
	h.controller.DeleteLesson(ctx, lessonDTO)
}
