package main

import "github.com/robfig/cron"

var Educations = [1]*education{&KBP}

func Launch() {
	for _, education := range Educations {
		c := cron.New()

		updateSchedule := func() {
			education.availableTypes = education.scheduleAvailableTypeUpdate()
			var subjects []Subject
			for _, availableType := range education.availableTypes {
				subjects = append(subjects, education.scheduleUpdate(availableType, education.states)...)
			}
		}

		primaryCron := cron.New()
		err := primaryCron.AddFunc(education.primaryScheduleUpdateCronPattern, func() {
			for i, state := range education.scheduleStatesUpdate(education.availableTypes[0]) {
				if state != education.states[i] {
					updateSchedule()
				}
			}
		})
		if err != nil {
			checkError(err, false)
			continue
		}
		err = c.AddFunc(education.scheduleUpdateCronPattern, updateSchedule)
		if err != nil {
			checkError(err, false)
			continue
		}
		err = c.AddFunc(education.primaryCronStartTimePattern, primaryCron.Start)
		if err != nil {
			checkError(err, false)
			continue
		}
		c.Start()
	}
}
