package handler

import (
	"context"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"studyum/internal/parser/apps"
	"studyum/internal/parser/controller"
	"studyum/internal/parser/dto"
	"time"
)

type Handler interface {
	Update(app apps.App)

	AddMark(mark dto.MarkDTO)
	EditMark(mark dto.MarkDTO)
	DeleteMark(mark primitive.ObjectID, studyPlaceID primitive.ObjectID)

	AddLesson(lesson dto.LessonDTO)
	EditLesson(lesson dto.LessonDTO)
	DeleteLesson(lesson dto.LessonDTO)
}

type handler struct {
	controller controller.Controller
}

func NewParserHandler(controller controller.Controller) Handler {
	h := &handler{controller: controller}

	for _, app := range controller.Apps() {
		ctx := context.Background()

		err, date := h.controller.GetLastUpdatedDate(ctx, app.StudyPlaceId())
		if err != nil {
			date = time.Now()
		}

		client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL")))
		if err != nil {
			logrus.Fatal(err)
		}

		if err = client.Connect(ctx); err != nil {
			logrus.Fatalf("Can't connect to database, error: %s", err.Error())
		}

		app.Init(date, client)

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

func (h *handler) Update(app apps.App) {
	ctx := context.Background()

	if !app.LaunchCron() {
		return
	}

	go h.controller.UpdateSchedule(ctx, app)
	go h.controller.UpdateJournal(ctx, app)
}

func (h *handler) AddMark(mark dto.MarkDTO) {
	ctx := context.Background()
	h.controller.AddMark(ctx, mark)
}

func (h *handler) EditMark(mark dto.MarkDTO) {
	ctx := context.Background()
	h.controller.EditMark(ctx, mark)
}

func (h *handler) DeleteMark(id primitive.ObjectID, studyPlaceID primitive.ObjectID) {
	ctx := context.Background()

	h.controller.DeleteMark(ctx, id, studyPlaceID)
}

func (h *handler) AddLesson(lesson dto.LessonDTO) {
	ctx := context.Background()

	h.controller.AddLesson(ctx, lesson)
}

func (h *handler) EditLesson(lesson dto.LessonDTO) {
	ctx := context.Background()

	h.controller.EditLesson(ctx, lesson)
}

func (h *handler) DeleteLesson(lesson dto.LessonDTO) {
	ctx := context.Background()

	h.controller.DeleteLesson(ctx, lesson)
}
