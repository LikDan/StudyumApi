package models

import "time"

type ScheduleInfo struct {
	Type          string     `json:"type" bson:"type"`
	TypeName      string     `json:"typeName" bson:"typeName"`
	StudyPlace    StudyPlace `json:"studyPlace" bson:"studyPlace"`
	StartWeekDate time.Time  `json:"startWeekDate" bson:"startWeekDate"`
	Date          time.Time  `json:"date" bson:"date"`
}
