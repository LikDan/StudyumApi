package models

type Journal struct {
	Info  JournalInfo  `json:"info" bson:"info"`
	Rows  []JournalRow `json:"rows" bson:"rows"`
	Dates []Lesson     `json:"dates" bson:"dates"`
}
