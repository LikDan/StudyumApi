package schedule

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Id           string    `json:"id" bson:"_id"`
	StudyPlaceId int       `json:"studyPlaceId" bson:"educationPlaceId"`
	Updated      bool      `json:"updated" bson:"updated"`
	Type         string    `json:"type" bson:"type"`
	EndDate      time.Time `json:"endDate" bson:"endDate"`
	StartDate    time.Time `json:"startDate" bson:"startDate"`
	Subject      string    `json:"subject" bson:"subject"`
	Group        string    `json:"group" bson:"group"`
	Teacher      string    `json:"teacher" bson:"teacher"`
	Room         string    `json:"room" bson:"room"`
	Title        string    `json:"title" bson:"smalldescription"`
	Homework     string    `json:"homework" bson:"homework"`
	Description  string    `json:"description" bson:"description"`
}

type GeneralLesson struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	StudyPlaceId int                `json:"studyPlaceId" bson:"studyPlaceId"`
	EndTime      string             `json:"endTime" bson:"endTime"`
	StartTime    string             `json:"startTime" bson:"startTime"`
	Subject      string             `json:"subject" bson:"subject"`
	Group        string             `json:"group" bson:"group"`
	Teacher      string             `json:"teacher" bson:"teacher"`
	Room         string             `json:"room" bson:"room"`
	DayIndex     int                `json:"dayIndex" bson:"dayIndex"`
	WeekIndex    int                `json:"weekIndex" bson:"weekIndex"`
}
