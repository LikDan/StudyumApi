package parser

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	h "studyium/api"
	"studyium/api/parser/app"
	"studyium/api/parser/studyPlace"
	"studyium/api/schedule"
	"studyium/db"
	"studyium/firebase"
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

	_, err := db.SubjectsCollection.InsertMany(nil, h.ToInterfaceSlice(subjects))
	h.CheckError(err, h.WARNING)
	_, err = db.StateCollection.DeleteMany(nil, bson.M{"educationPlaceId": edu.Id})
	h.CheckError(err, h.WARNING)
	_, err = db.StateCollection.InsertMany(nil, h.ToInterfaceSlice(edu.States))
	h.CheckError(err, h.WARNING)

	edu.LastUpdateTime = time.Now()

	if send {
		logrus.Info("Updated with notification")
		firebase.SendNotification("schedule_update", "Schedule", "Schedule was updated", "")

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

		find, err := db.StateCollection.Find(
			nil,
			bson.M{"educationPlaceId": edu.Id},
			options.Find().SetSort(bson.D{{"weekIndex", 1}, {"dayIndex", 1}}),
		)
		h.CheckError(err, h.WARNING)

		var states []schedule.StateInfo

		for find.TryNext(nil) {
			weekIndex := int(find.Current.Lookup("weekIndex").Int32())
			dayIndex := int(find.Current.Lookup("dayIndex").Int32())
			educationPlaceId := int(find.Current.Lookup("educationPlaceId").Int32())
			status := find.Current.Lookup("status").StringValue()

			state := schedule.StateInfo{
				State:        schedule.State(status),
				WeekIndex:    weekIndex,
				DayIndex:     dayIndex,
				StudyPlaceId: educationPlaceId,
			}

			states = append(states, state)
		}

		edu.States = states

		edu.GeneralCron = cron.New()
		edu.PrimaryCron = cron.New()

		edu.PrimaryCron.AddFunc(edu.PrimaryScheduleUpdateCronPattern, func() {
			if !h.EqualStateInfo(edu.States, edu.ScheduleStatesUpdate(edu.AvailableTypes[0])) {
				logrus.Info("Updated from primary cron (and stop it)")
				UpdateDbSchedule(edu)
				edu.PrimaryCron.Stop()
			} else {
				logrus.Info("No updates at primary cron")
			}
		})
		edu.GeneralCron.AddFunc(edu.ScheduleUpdateCronPattern, func() {
			UpdateDbSchedule(edu)
		})
		edu.GeneralCron.AddFunc(edu.PrimaryCronStartTimePattern, func() {
			if !edu.LaunchPrimaryCron {
				return
			}

			logrus.Info("Start primary cron")
			edu.PrimaryCron.Start()
		})
		edu.GeneralCron.Start()

		var generalSubjects []schedule.SubjectFull

		for _, availableType := range edu.AvailableTypes {
			generalSubjects = append(generalSubjects, edu.ScheduleUpdate(availableType, edu.States, edu.States, true)...)
		}

		_, err = db.GeneralSubjectsCollection.DeleteMany(nil, bson.D{{"educationPlaceId", edu.Id}})
		if h.CheckError(err, h.WARNING) {
			continue
		}
		_, err = db.GeneralSubjectsCollection.InsertMany(nil, h.ToInterfaceSlice(generalSubjects))
		if h.CheckError(err, h.WARNING) {
			continue
		}
	}
}
