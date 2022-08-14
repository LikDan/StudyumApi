package models

import (
	"studyum/src/utils"
	"time"
)

type Shift struct {
	Start time.Duration
	End   time.Duration
}

func BindShift(sHour, sMinute, eHour, eMinute int) Shift {
	return Shift{
		Start: utils.GetTimeDuration(sHour, sMinute),
		End:   utils.GetTimeDuration(eHour, eMinute),
	}
}
