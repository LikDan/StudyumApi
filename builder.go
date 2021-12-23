package main

import (
	"fmt"
	"github.com/robfig/cron"
	"strconv"
)

var Educations = [1]*education{&KBP}

func Launch() {
	for _, edu := range Educations {
		edu.availableTypes = edu.scheduleAvailableTypeUpdate()
		if len(edu.availableTypes) <= 0 {
			fmt.Printf("edu place with id: %s wasn't launched", strconv.Itoa(edu.educationPlaceId))
			continue
		}

		c := cron.New()

		updateSchedule := func() {
			edu.availableTypes = edu.scheduleAvailableTypeUpdate()
			edu.states = edu.scheduleStatesUpdate(edu.availableTypes[0])
			var subjects []Subject
			for _, availableType := range edu.availableTypes {
				subjects = append(subjects, edu.scheduleUpdate(availableType, edu.states)...)
			}
		}

		primaryCron := cron.New()
		err := primaryCron.AddFunc(edu.primaryScheduleUpdateCronPattern, func() {
			for i, state := range edu.scheduleStatesUpdate(edu.availableTypes[0]) {
				if len(edu.states) <= i || state != edu.states[i] {
					updateSchedule()
					sendNotification("schedule_update", "Schedule", "Schedule was updated", "")
					primaryCron.Stop()
				}
			}
		})
		if err != nil {
			checkError(err, false)
			continue
		}
		err = c.AddFunc(edu.scheduleUpdateCronPattern, updateSchedule)
		if err != nil {
			checkError(err, false)
			continue
		}
		primaryCron.Start()
		err = c.AddFunc(edu.primaryCronStartTimePattern, primaryCron.Start)
		if err != nil {
			checkError(err, false)
			continue
		}
		c.Start()
	}
}
