package parser

import (
	"fmt"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
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
	return //only for dev
	lastStates := edu.States
	send := !h.EqualStateInfo(edu.States, edu.ScheduleStatesUpdate(edu.AvailableTypes[0]))
	edu.AvailableTypes = edu.ScheduleAvailableTypeUpdate()
	edu.States = edu.ScheduleStatesUpdate(edu.AvailableTypes[0])
	var subjects []schedule.SubjectFull

	for _, availableType := range edu.AvailableTypes {
		subjects = append(subjects, edu.ScheduleUpdate(availableType, edu.States, lastStates, false)...)
	}

	_, err := db.SubjectsCollection.InsertMany(nil, h.ToInterfaceSlice(subjects))
	h.CheckError(err)
	_, err = db.StateCollection.DeleteMany(nil, bson.M{"educationPlaceId": edu.Id})
	h.CheckError(err)
	_, err = db.StateCollection.InsertMany(nil, h.ToInterfaceSlice(edu.States))
	h.CheckError(err)

	edu.LastUpdateTime = time.Now()

	if send {
		lastStatesString := ""
		currentStatesString := ""

		for _, state := range lastStates {
			lastStatesString += state.ToJsonWithoutId()
		}

		for _, state := range edu.States {
			currentStatesString += state.ToJsonWithoutId()
		}

		log.Printf("Schedule updated from\n" + lastStatesString + "\nto\n" + currentStatesString)

		firebase.SendNotification("schedule_update", "Schedule", "Schedule was updated", "")
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
		h.CheckError(err)

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
				UpdateDbSchedule(edu)
				edu.PrimaryCron.Stop()
			} else {
				log.Println("No updates")
			}
		})
		edu.GeneralCron.AddFunc(edu.ScheduleUpdateCronPattern, func() {
			UpdateDbSchedule(edu)
		})
		edu.GeneralCron.AddFunc(edu.PrimaryCronStartTimePattern, func() {
			if !edu.LaunchPrimaryCron {
				return
			}

			edu.PrimaryCron.Start()
		})
		edu.GeneralCron.Start()

		var generalSubjects []schedule.SubjectFull

		for _, availableType := range edu.AvailableTypes {
			generalSubjects = append(generalSubjects, edu.ScheduleUpdate(availableType, edu.States, edu.States, true)...)
		}

		_, err = db.GeneralSubjectsCollection.DeleteMany(nil, bson.D{{"educationPlaceId", edu.Id}})
		if h.CheckError(err) {
			continue
		}
		_, err = db.GeneralSubjectsCollection.InsertMany(nil, h.ToInterfaceSlice(generalSubjects))
		if h.CheckError(err) {
			continue
		}
	}
}
