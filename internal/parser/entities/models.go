package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ParsedInfoType map[string]any

type State string

const (
	Updated    State = "UPDATED"
	NotUpdated State = "NOT_UPDATED"
)

type DayState struct {
	State        State              `json:"status"`
	WeekIndex    int                `json:"weekIndex"`
	DayIndex     int                `json:"dayIndex"`
	StudyPlaceId primitive.ObjectID `json:"-"`
}

type JournalUser struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	ParserAppName string             `bson:"parserAppName" json:"parserAppName"`
	Login         string             `bson:"login" json:"login"`
	Password      string             `bson:"password" json:"password"`
	AdditionInfo  map[string]string  `bson:"additionInfo" json:"additionInfo"`
}

type ScheduleTypeInfo struct {
	Id            primitive.ObjectID `bson:"_id" json:"id"`
	ParserAppName string             `bson:"parserAppName" json:"parserAppName"`
	Group         string             `bson:"group" json:"group"`
	Url           string             `bson:"url" json:"url"`
}

type ScheduleStateInfo struct {
	State     State `json:"status"`
	WeekIndex int   `json:"weekIndex"`
	DayIndex  int   `json:"dayIndex"`
}

func GetScheduleStateInfoByIndexes(states []ScheduleStateInfo, weekIndex, dayIndex int) ScheduleStateInfo {
	for _, state := range states {
		if state.WeekIndex == weekIndex && state.DayIndex == dayIndex {
			return state
		}
	}

	return ScheduleStateInfo{
		State:     NotUpdated,
		WeekIndex: weekIndex,
		DayIndex:  dayIndex,
	}
}

type Shift struct {
	Start time.Duration
	End   time.Duration
	Date  time.Time
}

func NewShift(sHour, sMinute, eHour, eMinute int) Shift {
	return Shift{
		Start: time.Duration(sHour*60*60+sMinute*60) * time.Second,
		End:   time.Duration(eHour*60*60+eMinute*60) * time.Second,
	}
}
