package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
)

type State string

const (
	Updated    State = "UPDATED"
	NotUpdated State = "NOT_UPDATED"
)

type StateInfo struct {
	State        State `json:"state"`
	WeekIndex    int   `json:"weekIndex"`
	DayIndex     int   `json:"dayIndex"`
	StudyPlaceId int   `json:"-"`
}

func stateToBson(info StateInfo) bson.D {
	return bson.D{
		{"weekIndex", info.WeekIndex},
		{"dayIndex", info.DayIndex},
		{"status", info.State},
		{"educationPlaceId", info.StudyPlaceId},
	}
}

func (s StateInfo) toJsonWithoutId() string {
	return "{\"status\": \"" + string(s.State) +
		"\", \"weekIndex\": " + strconv.Itoa(s.WeekIndex) +
		", \"dayIndex\": " + strconv.Itoa(s.DayIndex) + "}"
}
