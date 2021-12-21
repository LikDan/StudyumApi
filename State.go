package main

import "go.mongodb.org/mongo-driver/bson"

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
