package main

import (
	"fmt"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"time"
)

var Educations = [1]*education{&KBP}

func UpdateDbSchedule(edu *education) {
	lastStates := edu.States
	send := !EqualStateInfo(edu.States, edu.scheduleStatesUpdate(edu.AvailableTypes[0]))
	edu.AvailableTypes = edu.scheduleAvailableTypeUpdate()
	edu.States = edu.scheduleStatesUpdate(edu.AvailableTypes[0])
	var subjects []SubjectFull

	for _, availableType := range edu.AvailableTypes {
		subjects = append(subjects, edu.scheduleUpdate(availableType, edu.States, lastStates, false)...)
	}

	_, err := subjectsCollection.InsertMany(nil, ToInterfaceSlice(subjects))
	checkError(err)
	_, err = stateCollection.DeleteMany(nil, bson.M{"educationPlaceId": edu.id})
	checkError(err)
	_, err = stateCollection.InsertMany(nil, ToInterfaceSlice(edu.States))
	checkError(err)

	edu.LastUpdateTime = time.Now()

	if send {
		lastStatesString := ""
		currentStatesString := ""

		for _, state := range lastStates {
			lastStatesString += state.toJsonWithoutId()
		}

		for _, state := range edu.States {
			currentStatesString += state.toJsonWithoutId()
		}

		log.Printf("Schedule updated from\n" + lastStatesString + "\nto\n" + currentStatesString)

		sendNotification("schedule_update", "Schedule", "Schedule was updated", "")
	}
}

func Launch() {
	for _, edu := range Educations {
		edu.AvailableTypes = edu.scheduleAvailableTypeUpdate()
		if len(edu.AvailableTypes) <= 0 {
			fmt.Printf("edu place with id: %s wasn't launched\n", strconv.Itoa(edu.id))
			continue
		}

		find, err := stateCollection.Find(
			nil,
			bson.M{"educationPlaceId": edu.id},
			options.Find().SetSort(bson.D{{"weekIndex", 1}, {"dayIndex", 1}}),
		)
		checkError(err)

		var states []StateInfo

		for find.TryNext(nil) {
			weekIndex := int(find.Current.Lookup("weekIndex").Int32())
			dayIndex := int(find.Current.Lookup("dayIndex").Int32())
			educationPlaceId := int(find.Current.Lookup("educationPlaceId").Int32())
			status := find.Current.Lookup("status").StringValue()

			state := StateInfo{
				State:        State(status),
				WeekIndex:    weekIndex,
				DayIndex:     dayIndex,
				StudyPlaceId: educationPlaceId,
			}

			states = append(states, state)
		}

		edu.States = states

		edu.generalCron = cron.New()
		edu.primaryCron = cron.New()

		edu.primaryCron.AddFunc(edu.PrimaryScheduleUpdateCronPattern, func() {
			if !EqualStateInfo(edu.States, edu.scheduleStatesUpdate(edu.AvailableTypes[0])) {
				UpdateDbSchedule(edu)
				edu.primaryCron.Stop()
			} else {
				log.Println("No updates")
			}
		})
		edu.generalCron.AddFunc(edu.ScheduleUpdateCronPattern, func() {
			UpdateDbSchedule(edu)
		})
		edu.generalCron.AddFunc(edu.PrimaryCronStartTimePattern, func() {
			if !edu.LaunchPrimaryCron {
				return
			}

			edu.primaryCron.Start()
		})
		edu.generalCron.Start()

		var generalSubjects []SubjectFull

		for _, availableType := range edu.AvailableTypes {
			generalSubjects = append(generalSubjects, edu.scheduleUpdate(availableType, edu.States, edu.States, true)...)
		}

		_, err = generalSubjectsCollection.DeleteMany(nil, bson.D{{"educationPlaceId", edu.id}})
		if checkError(err) {
			continue
		}
		_, err = generalSubjectsCollection.InsertMany(nil, ToInterfaceSlice(generalSubjects))
		if checkError(err) {
			continue
		}
	}
}
