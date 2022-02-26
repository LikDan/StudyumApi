package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type SubjectFull struct {
	Subject          string    `json:"subject"`
	Teacher          string    `json:"teacher"`
	Group            string    `json:"group"`
	Room             string    `json:"room"`
	ColumnIndex      int       `json:"columnIndex"`
	RowIndex         int       `json:"rowIndex"`
	WeekIndex        int       `json:"weekIndex"`
	Type_            string    `json:"type" bson:"type"`
	EducationPlaceId int       `json:"educationPlaceId"`
	Date             time.Time `json:"date"`
}

func subjectToBson(subject SubjectFull) bson.D {
	return bson.D{
		{"date", subject.Date},
		{"columnIndex", subject.ColumnIndex},
		{"rowIndex", subject.RowIndex},
		{"weekIndex", subject.WeekIndex},
		{"subject", subject.Subject},
		{"teacher", subject.Teacher},
		{"group", subject.Group},
		{"room", subject.Room},
		{"type", subject.Type_},
		{"educationPlaceId", subject.EducationPlaceId},
	}
}

func subjectToBsonWithoutType(subject SubjectFull) bson.D {
	return bson.D{
		{"columnIndex", subject.ColumnIndex},
		{"rowIndex", subject.RowIndex},
		{"weekIndex", subject.WeekIndex},
		{"subject", subject.Subject},
		{"teacher", subject.Teacher},
		{"group", subject.Group},
		{"room", subject.Room},
		{"educationPlaceId", subject.EducationPlaceId},
	}
}
