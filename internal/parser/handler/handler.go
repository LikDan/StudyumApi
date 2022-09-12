package handler

import (
	"context"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
	"studyum/internal/parser/apps"
	"studyum/internal/parser/controller"
	"studyum/internal/parser/dto"
	"time"
)

type Handler interface {
	Update(app apps.App)

	AddMark(mark entities.Mark)
	EditMark(mark entities.Mark)
	DeleteMark(mark entities.Mark)

	AddLesson(lesson entities.Lesson)
	EditLesson(lesson entities.Lesson)
	DeleteLesson(lesson entities.Lesson)

	GetSignUpDataByCode(ctx context.Context, code string) (entities.SignUpCode, error)
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

		app.Init(date)

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

func (h *handler) AddMark(mark entities.Mark) {
	ctx := context.Background()

	markDTO := dto.MarkDTO{
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

	markDTO := dto.MarkDTO{
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

	markDTO := dto.MarkDTO{
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

	lessonDTO := dto.LessonDTO{
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

	lessonDTO := dto.LessonDTO{
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

	lessonDTO := dto.LessonDTO{
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

func (h *handler) GetSignUpDataByCode(ctx context.Context, code string) (entities.SignUpCode, error) {
	codeDTO, err := h.controller.GetSignUpDataByCode(ctx, code)
	if err != nil {
		return entities.SignUpCode{}, err
	}

	codeData := entities.SignUpCode{
		Id:           primitive.NilObjectID,
		Code:         codeDTO.Code,
		Name:         codeDTO.Name,
		StudyPlaceID: codeDTO.StudyPlaceID,
		Type:         codeDTO.Type,
		Typename:     codeDTO.Typename,
	}

	return codeData, nil
}
