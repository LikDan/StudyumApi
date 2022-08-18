package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type State string

const (
	Updated    State = "UPDATED"
	NotUpdated State = "NOT_UPDATED"
)

type DayState struct {
	State        State `bson:"status" json:"status"`
	WeekIndex    int   `bson:"weekIndex" json:"weekIndex"`
	DayIndex     int   `bson:"dayIndex" json:"dayIndex"`
	StudyPlaceId int   `bson:"educationPlaceId" json:"-"`
}

type JournalUser struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	ParserAppName  string             `bson:"parserAppName" json:"parserAppName"`
	Login          string             `bson:"login" json:"login"`
	Password       string             `bson:"password" json:"password"`
	AdditionInfo   map[string]string  `bson:"additionInfo" json:"additionInfo"`
	LastParsedDate time.Time          `bson:"lastParsedDate" json:"lastParsedDate"`
}

type ScheduleTypeInfo struct {
	Id            primitive.ObjectID `bson:"_id" json:"id"`
	ParserAppName string             `bson:"parserAppName" json:"parserAppName"`
	Group         string             `bson:"group" json:"group"`
	Url           string             `bson:"url" json:"url"`
}

type ScheduleStateInfo struct {
	State     State `bson:"status" json:"status"`
	WeekIndex int   `bson:"weekIndex" json:"weekIndex"`
	DayIndex  int   `bson:"dayIndex" json:"dayIndex"`
}

func GetScheduleStateInfoByIndexes(weekIndex, dayIndex int, states []ScheduleStateInfo) ScheduleStateInfo {
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
