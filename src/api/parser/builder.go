package parser

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	h "studyum/src/api"
	"studyum/src/api/parser/app"
	"studyum/src/api/parser/studyPlace"
	"studyum/src/api/schedule"
	"studyum/src/db"
	"studyum/src/models"
	"studyum/src/utils"
	"time"
)

var Educations = [1]*studyPlace.Education{&app.KBP}

func UpdateDbSchedule(edu *studyPlace.Education) {
	logrus.Infof("Update schedule for %s", edu.Name)
	lastStates := edu.States
	send := !h.EqualStateInfo(edu.States, edu.ScheduleStatesUpdate(edu.AvailableTypes[0]))
	edu.AvailableTypes = edu.ScheduleAvailableTypeUpdate()
	edu.States = edu.ScheduleStatesUpdate(edu.AvailableTypes[0])
	var subjects []schedule.SubjectFull

	for _, availableType := range edu.AvailableTypes {
		subjects = append(subjects, edu.ScheduleUpdate(availableType, edu.States, lastStates, false)...)
	}

	_, err := db.LessonsCollection.InsertMany(nil, h.ToInterfaceSlice(subjects))
	h.CheckError(err, h.WARNING)
	_, err = db.StateCollection.DeleteMany(nil, bson.M{"educationPlaceId": edu.Id})
	h.CheckError(err, h.WARNING)
	_, err = db.StateCollection.InsertMany(nil, h.ToInterfaceSlice(edu.States))
	h.CheckError(err, h.WARNING)

	edu.LastUpdateTime = time.Now()

	if send {
		logrus.Info("Updated with notification")
		utils.SendNotification("schedule_update", "Schedule", "Schedule was updated", "")

		lastStatesBytes, err := json.Marshal(lastStates)
		if h.CheckError(err, h.WARNING) {
			return
		}
		currentStatesBytes, err := json.Marshal(edu.States)
		if h.CheckError(err, h.WARNING) {
			return
		}

		logrus.Info("Schedule was updated")
		logrus.Info("{\"lastStates\": \"" + string(lastStatesBytes) + "\", \"currentStates\": \"" + string(currentStatesBytes) + "\"}")
	}
}

func Launch() {
	for _, edu := range Educations {
		edu.AvailableTypes = edu.ScheduleAvailableTypeUpdate()
		if len(edu.AvailableTypes) <= 0 {
			fmt.Printf("edu place with id: %s wasn't launched\n", strconv.Itoa(edu.Id))
			continue
		}

		cursor, err := db.StateCollection.Find(
			nil,
			bson.M{"educationPlaceId": edu.Id},
			options.Find().SetSort(bson.D{{"weekIndex", 1}, {"dayIndex", 1}}),
		)
		if h.CheckError(err, h.WARNING) {
			return
		}

		var states []schedule.StateInfo
		if err := cursor.All(nil, &states); h.CheckError(err, h.WARNING) {
			return
		}

		edu.States = states

		edu.GeneralCron = cron.New()
		edu.GeneralCron.AddFunc(edu.ScheduleUpdateCronPattern, func() {
			UpdateDbSchedule(edu)
		})

		edu.GeneralCron.Start()

		//UpdateGeneral(edu)
	}
}

func UpdateGeneral(edu *studyPlace.Education) {
	var generalSubjectsRaw []schedule.SubjectFull

	for _, availableType := range edu.AvailableTypes {
		generalSubjectsRaw = append(generalSubjectsRaw, edu.ScheduleUpdate(availableType, edu.States, edu.States, true)...)
	}

	var generalSubjects []models.GeneralLesson
	for _, lessonRaw := range generalSubjectsRaw {
		lesson := models.GeneralLesson{
			Id:           lessonRaw.Id,
			StudyPlaceId: lessonRaw.EducationPlaceId,
			EndTime:      lessonRaw.EndTime.Format("15:04"),
			StartTime:    lessonRaw.StartTime.Format("15:04"),
			Subject:      lessonRaw.Subject,
			Group:        lessonRaw.Group,
			Teacher:      lessonRaw.Teacher,
			Room:         lessonRaw.Room,
			DayIndex:     lessonRaw.ColumnIndex,
			WeekIndex:    lessonRaw.WeekIndex,
		}

		generalSubjects = append(generalSubjects, lesson)
	}

	_, err := db.GeneralLessonsCollection.DeleteMany(nil, bson.D{{"studyPlaceId", edu.Id}})
	if h.CheckError(err, h.WARNING) {
		return
	}
	_, err = db.GeneralLessonsCollection.InsertMany(nil, h.ToInterfaceSlice(generalSubjects))
	if h.CheckError(err, h.WARNING) {
		return
	}
}
