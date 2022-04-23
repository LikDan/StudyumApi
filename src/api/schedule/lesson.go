package schedule

import (
	"time"
)

type LessonOld struct {
	Id          int           `bson:"_id" json:"-"`
	Subjects    []*SubjectOld `bson:"subjects" json:"subjects"`
	ColumnIndex int32         `bson:"columnIndex" json:"columnIndex"`
	RowIndex    int32         `bson:"rowIndex" json:"rowIndex"`
	WeekIndex   int32         `bson:"weekIndex" json:"weekIndex"`
	Date        time.Time     `bson:"date" json:"-"`
	IsStay      bool          `bson:"isStay" json:"isStay"`
}

type SubjectOld struct {
	Subject string `bson:"subject" json:"subject"`
	Teacher string `bson:"teacher" json:"teacher"`
	Group   string `bson:"group" json:"group"`
	Room    string `bson:"room" json:"room"`
	Type_   string `bson:"type" json:"type"`
}

type Lesson struct {
	Subjects     []*Subject `bson:"subjects" json:"subjects"`
	StartDate    time.Time  `bson:"startDate" json:"startDate"`
	EndDate      time.Time  `bson:"endDate" json:"endDate"`
	Updated      bool       `bson:"updated" json:"updated"`
	StudyPlaceId int        `bson:"studyPlaceId" json:"studyPlaceId"`
}

type Subject struct {
	Subject     string `bson:"subject" json:"subject"`
	Teacher     string `bson:"teacher" json:"teacher"`
	Group       string `bson:"group" json:"group"`
	Room        string `bson:"room" json:"room"`
	Type_       string `bson:"type" json:"type"`
	Description string `bson:"description" json:"description"`
	Title       string `bson:"title" json:"title"`
	Homework    string `bson:"homework" json:"homework"`
}
