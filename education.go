package main

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

type education struct {
	id                               int
	scheduleUpdateCronPattern        string
	primaryScheduleUpdateCronPattern string
	primaryCronStartTimePattern      string
	scheduleUpdate                   func(string, []StateInfo, bool) []SubjectFull
	scheduleStatesUpdate             func(string) []StateInfo
	scheduleAvailableTypeUpdate      func() []string
	availableTypes                   []string
	states                           []StateInfo
	password                         string

	primaryCron       *cron.Cron
	generalCron       *cron.Cron
	launchPrimaryCron bool
}

func (e education) statesJSON() []gin.H {
	var statesJSON []gin.H
	for _, state := range e.states {
		stateJSON := gin.H{
			"weekIndex": state.WeekIndex,
			"dayIndex":  state.DayIndex,
			"state":     state.State,
		}

		statesJSON = append(statesJSON, stateJSON)
	}

	return statesJSON
}
