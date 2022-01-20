package main

import (
	"time"
)

type Lesson struct {
	Id          int        `bson:"_id" json:"-"`
	Subjects    []*Subject `bson:"subjects" json:"subjects"`
	ColumnIndex int32      `bson:"columnIndex" json:"columnIndex"`
	RowIndex    int32      `bson:"rowIndex" json:"rowIndex"`
	WeekIndex   int32      `bson:"weekIndex" json:"weekIndex"`
	Date        time.Time  `bson:"date" json:"-"`
	IsStay      bool       `bson:"isStay" json:"isStay"`
}

type Subject struct {
	Subject string `bson:"subject" json:"subject"`
	Teacher string `bson:"teacher" json:"teacher"`
	Group   string `bson:"group" json:"group"`
	Room    string `bson:"room" json:"room"`
	Type_   string `bson:"type" json:"type"`
}
