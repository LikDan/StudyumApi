package handler

import (
	"context"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"studyum/internal/parser/application"
	"studyum/internal/parser/controller"
	"studyum/pkg/firebase"
)

type Handler interface {
	Update(app application.App)
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
