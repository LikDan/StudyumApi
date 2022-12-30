package dto

import "time"

type MarksReport struct {
	LessonType string     `json:"lessonType" bson:"lessonType"`
	Mark       string     `json:"mark" bson:"mark"`
	StartDate  *time.Time `json:"startDate" bson:"startDate"`
	EndDate    *time.Time `json:"endDate" bson:"endDate"`
	NotExists  bool       `json:"notExists" bson:"notExists"`
}

type AbsencesReport struct {
	StartDate *time.Time `json:"startDate" bson:"startDate"`
	EndDate   *time.Time `json:"endDate" bson:"endDate"`
}
