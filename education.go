package main

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"time"
)

type education struct {
	id                               int
	ScheduleUpdateCronPattern        string
	PrimaryScheduleUpdateCronPattern string
	PrimaryCronStartTimePattern      string
	scheduleUpdate                   func(string, []StateInfo, bool) []SubjectFull
	scheduleStatesUpdate             func(string) []StateInfo
	scheduleAvailableTypeUpdate      func() []string
	AvailableTypes                   []string
	States                           []StateInfo
	password                         string

	primaryCron       *cron.Cron
	generalCron       *cron.Cron
	LaunchPrimaryCron bool
	LastUpdateTime    time.Time
}

func (e education) statesJSON() []gin.H {
	var statesJSON []gin.H
	for _, state := range e.States {
		stateJSON := gin.H{
			"weekIndex": state.WeekIndex,
			"dayIndex":  state.DayIndex,
			"state":     state.State,
		}

		statesJSON = append(statesJSON, stateJSON)
	}

	return statesJSON
}
