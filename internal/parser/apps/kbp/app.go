package kbp

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/parser/appDTO"
	"studyum/internal/parser/apps"
	"studyum/internal/parser/apps/kbp/controller"
	"studyum/internal/parser/apps/kbp/repository"
	"studyum/internal/parser/entities"
	"studyum/pkg/datetime"
	"time"
)

type app struct {
	States     []entities.ScheduleStateInfo
	TempStates []entities.ScheduleStateInfo

	WeekdaysShift []entities.Shift
	WeekendsShift []entities.Shift

	DefaultColor string
	AddedColor   string
	RemovedColor string

	controller *controller.Controller
}

func NewApp() apps.App {
	weekdaysShift := []entities.Shift{
		entities.NewShift(8, 00, 9, 35),
		entities.NewShift(9, 45, 11, 20),
		entities.NewShift(11, 50, 13, 25),
		entities.NewShift(13, 45, 15, 20),
		entities.NewShift(15, 40, 17, 15),
		entities.NewShift(17, 25, 19, 0),
		entities.NewShift(19, 10, 20, 45),
	}

	weekendsShift := []entities.Shift{
		entities.NewShift(8, 00, 9, 35),
		entities.NewShift(9, 45, 11, 20),
		entities.NewShift(11, 30, 13, 5),
		entities.NewShift(13, 30, 15, 5),
		entities.NewShift(15, 15, 16, 50),
		entities.NewShift(17, 0, 18, 35),
		entities.NewShift(18, 45, 20, 20),
	}

	return &app{
		WeekdaysShift: weekdaysShift,
		WeekendsShift: weekendsShift,

		DefaultColor: "#F1F1F1",
		AddedColor:   "#71AB7F",
		RemovedColor: "#FA6F46",
	}
}

func (a *app) Init(date time.Time, client *mongo.Client) {
	states := make([]entities.ScheduleStateInfo, 14)

	dateCursor := datetime.Date().AddDate(0, 0, int(datetime.Date().Weekday()))
	for i := 0; i < 14; i++ {
		state := entities.ScheduleStateInfo{
			WeekIndex: i / 7,
			DayIndex:  i % 7,
		}
		if date.Before(dateCursor) {
			state.State = entities.NotUpdated
		} else {
			state.State = entities.Updated
		}

		states[i] = state

		dateCursor.AddDate(0, 0, 1)
	}

	a.States = states

	repo := repository.NewRepository(client)
	a.controller = controller.NewController(repo)
}

func (a *app) GetName() string              { return "kbp" }
func (a *app) GetUpdateCronPattern() string { return "@every 30m" }
func (a *app) LaunchCron() bool             { return false }
func (a *app) StudyPlaceId() primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex("631261e11b8b855cc75cec35")
	return id
}

func (a *app) ScheduleUpdate(entities.ScheduleTypeInfo) []appDTO.LessonDTO {
	return nil
}

func (a *app) GeneralScheduleUpdate(entities.ScheduleTypeInfo) []appDTO.GeneralLessonDTO {
	return nil
}

func (a *app) ScheduleTypesUpdate() []appDTO.ScheduleTypeInfoDTO {
	return nil
}

func (a *app) JournalUpdate(entities.JournalUser) []appDTO.MarkDTO {
	return nil
}

func (a *app) CommitUpdate() {
}
