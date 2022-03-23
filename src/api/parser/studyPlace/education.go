package studyPlace

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"studyium/api/schedule"
	"time"
)

type Education struct {
	Id                               int
	Name                             string
	ScheduleUpdateCronPattern        string
	PrimaryScheduleUpdateCronPattern string
	PrimaryCronStartTimePattern      string
	ScheduleUpdate                   func(string, []schedule.StateInfo, []schedule.StateInfo, bool) []schedule.SubjectFull `json:"-"`
	ScheduleStatesUpdate             func(string) []schedule.StateInfo                                                     `json:"-"`
	ScheduleAvailableTypeUpdate      func() []string                                                                       `json:"-"`
	AvailableTypes                   []string
	States                           []schedule.StateInfo
	Password                         string `json:"-"`

	PrimaryCron       *cron.Cron
	GeneralCron       *cron.Cron
	LaunchPrimaryCron bool
	LastUpdateTime    time.Time
}

func (e Education) statesJSON() []gin.H {
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
