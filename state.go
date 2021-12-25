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
	state            State
	weekIndex        int
	dayIndex         int
	educationPlaceId int
}

func stateToBson(info StateInfo) bson.D {
	return bson.D{
		{"weekIndex", info.weekIndex},
		{"dayIndex", info.dayIndex},
		{"status", info.state},
		{"educationPlaceId", info.educationPlaceId},
	}
}

func (s StateInfo) toJsonWithoutId() string {
	return "{\"status\": \"" + string(s.state) +
		"\", \"weekIndex\": " + strconv.Itoa(s.weekIndex) +
		", \"dayIndex\": " + strconv.Itoa(s.dayIndex) + "}"
}
