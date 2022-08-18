package handler

import (
	"context"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"studyum/internal/parser/controller"
	"studyum/internal/parser/entities"
	"studyum/pkg/firebase"
)

type Handler struct {
	firebase firebase.Firebase

	controller controller.IController
}

func NewParserHandler(firebase firebase.Firebase, controller controller.IController) *Handler {
	handler := &Handler{firebase: firebase, controller: controller}

	for _, app := range controller.Apps() {
		ctx := context.Background()
		lastLesson, _ := handler.controller.GetLastLesson(ctx, app.StudyPlaceId())

		app.Init(lastLesson)

		types := app.ScheduleTypesUpdate()
		if err := handler.controller.InsertScheduleTypes(ctx, types); err != nil {
			continue
		}

		updateCron := cron.New()
		if err := updateCron.AddFunc(app.GetUpdateCronPattern(), func() { handler.Update(app) }); err != nil {
			logrus.Warningf("cannot launch cron for %s, err: %e", app.GetName(), err)
			continue
		}

		updateCron.Start()
	}

	return handler
}

func (h *Handler) Update(app entities.IApp) {
	ctx := context.Background()
	h.controller.Update(ctx, app)
}
