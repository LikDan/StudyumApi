package models

type Schedule struct {
	Info    ScheduleInfo `json:"info" bson:"info"`
	Lessons []*Lesson    `json:"lessons" bson:"lessons"`
}
