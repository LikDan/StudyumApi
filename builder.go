package main

import (
	"fmt"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
)

var Educations = [1]*education{&KBP}

func UpdateDbSchedule(edu *education) {
	lastStates := edu.states
	send := !EqualStateInfo(edu.states, edu.scheduleStatesUpdate(edu.availableTypes[0]))
	edu.availableTypes = edu.scheduleAvailableTypeUpdate()
	edu.states = edu.scheduleStatesUpdate(edu.availableTypes[0])
	var subjects []SubjectFull
	for _, availableType := range edu.availableTypes {
		subjects = append(subjects, edu.scheduleUpdate(availableType, edu.states, false)...)
	}
	var subjectsBSON []interface{}
	for _, subject := range subjects {
		subjectsBSON = append(subjectsBSON, subjectToBson(subject))
	}

	var stateBSON []interface{}
	for _, state := range edu.states {
		stateBSON = append(stateBSON, stateToBson(state))
	}

	_, err := subjectsCollection.DeleteMany(nil, bson.D{{"educationPlaceId", edu.id}})
	checkError(err)
	_, err = subjectsCollection.InsertMany(nil, subjectsBSON)
	checkError(err)
	_, err = stateCollection.DeleteMany(nil, bson.D{{"educationPlaceId", edu.id}})
	checkError(err)
	_, err = stateCollection.InsertMany(nil, stateBSON)
	checkError(err)

	if send {
		lastStatesString := ""
		currentStatesString := ""

		for _, state := range lastStates {
			lastStatesString += state.toJsonWithoutId()
		}

		for _, state := range edu.states {
			currentStatesString += state.toJsonWithoutId()
		}

		log.Printf("Schedule updated from\n" + lastStatesString + "\nto\n" + currentStatesString)

		sendNotification("schedule_update", "Schedule", "Schedule was updated", "")
	}
}

func Launch() {
	for _, edu := range Educations {
		edu.availableTypes = edu.scheduleAvailableTypeUpdate()
		if len(edu.availableTypes) <= 0 {
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

		edu.states = states

		edu.generalCron = cron.New()
		edu.primaryCron = cron.New()

		err = edu.primaryCron.AddFunc(edu.primaryScheduleUpdateCronPattern, func() {
			if !EqualStateInfo(edu.states, edu.scheduleStatesUpdate(edu.availableTypes[0])) {
				UpdateDbSchedule(edu)
				edu.primaryCron.Stop()
			} else {
				log.Println("No updates")
			}
		})
		if checkError(err) {
			continue
		}
		err = edu.generalCron.AddFunc(edu.scheduleUpdateCronPattern, func() {
			UpdateDbSchedule(edu)
		})
		if checkError(err) {
			continue
		}
		err = edu.generalCron.AddFunc(edu.primaryCronStartTimePattern, func() {
			if !edu.launchPrimaryCron {
				return
			}

			edu.primaryCron.Start()
		})
		if checkError(err) {
			continue
		}
		edu.generalCron.Start()

		var generalSubjects []SubjectFull

		for _, availableType := range edu.availableTypes {
			generalSubjects = append(generalSubjects, edu.scheduleUpdate(availableType, edu.states, true)...)
		}

		var generalSubjectsBson []interface{}
		for _, subject := range generalSubjects {
			generalSubjectsBson = append(generalSubjectsBson, subjectToBsonWithoutType(subject))
		}

		_, err = generalSubjectsCollection.DeleteMany(nil, bson.D{{"educationPlaceId", edu.id}})
		if checkError(err) {
			continue
		}
		_, err = generalSubjectsCollection.InsertMany(nil, generalSubjectsBson)
		if checkError(err) {
			continue
		}
	}
}
