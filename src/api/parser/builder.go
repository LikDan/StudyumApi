package parser

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	h "studyium/src/api"
	"studyium/src/api/parser/app"
	"studyium/src/api/parser/studyPlace"
	"studyium/src/api/schedule"
	"studyium/src/db"
	"studyium/src/firebase"
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
		edu.GeneralCron.AddFunc(edu.ScheduleUpdateCronPattern, func() {
			UpdateDbSchedule(edu)
		})

		edu.GeneralCron.Start()
	}
}
