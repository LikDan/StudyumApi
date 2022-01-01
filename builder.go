package main

import (
	"fmt"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
)

var Educations = [1]*education{&KBP}

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

		for find.TryNext(nil) {
			weekIndex := int(find.Current.Lookup("weekIndex").Int32())
			dayIndex := int(find.Current.Lookup("dayIndex").Int32())
			educationPlaceId := int(find.Current.Lookup("educationPlaceId").Int32())
			status := find.Current.Lookup("status").StringValue()

			state := StateInfo{
				state:            State(status),
				weekIndex:        weekIndex,
				dayIndex:         dayIndex,
				educationPlaceId: educationPlaceId,
			}

			edu.states = append(edu.states, state)
		}

		c := cron.New()
		primaryCron := cron.New()

		updateSchedule := func() {
			send := EqualStateInfo(edu.states, edu.scheduleStatesUpdate(edu.availableTypes[0]))
			edu.availableTypes = edu.scheduleAvailableTypeUpdate()
			edu.states = edu.scheduleStatesUpdate(edu.availableTypes[0])
			var subjects []SubjectFull
			for _, availableType := range edu.availableTypes {
				subjects = append(subjects, edu.scheduleUpdate(availableType, edu.states)...)
			}
			var subjectsBSON []interface{}
			for _, subject := range subjects {
				subjectsBSON = append(subjectsBSON, subjectToBson(subject))
			}

			var stateBSON []interface{}
			for _, state := range edu.states {
				stateBSON = append(stateBSON, stateToBson(state))
			}

			err := subjectsCollection.Drop(nil)
			checkError(err)
			_, err = subjectsCollection.InsertMany(nil, subjectsBSON)
			checkError(err)
			err = stateCollection.Drop(nil)
			checkError(err)
			_, err = stateCollection.InsertMany(nil, stateBSON)
			checkError(err)

			if send {
				sendNotification("schedule_update", "Schedule", "Schedule was updated", "")
			}
		}

		err = primaryCron.AddFunc(edu.primaryScheduleUpdateCronPattern, func() {
			if EqualStateInfo(edu.states, edu.scheduleStatesUpdate(edu.availableTypes[0])) {
				updateSchedule()
				primaryCron.Stop()
			}
		})
		if checkError(err) {
			continue
		}
		err = c.AddFunc(edu.scheduleUpdateCronPattern, updateSchedule)
		if checkError(err) {
			continue
		}
		primaryCron.Start()
		err = c.AddFunc(edu.primaryCronStartTimePattern, primaryCron.Start)
		if checkError(err) {
			continue
		}
		c.Start()
	}
}
