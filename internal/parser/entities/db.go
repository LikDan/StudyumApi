package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type GeneralLesson struct {
	Id           primitive.ObjectID `bson:"_id"`
	StudyPlaceId int                `bson:"studyPlaceId"`
	EndTime      string             `bson:"endTime"`
	StartTime    string             `bson:"startTime"`
	Subject      string             `bson:"subject"`
	Group        string             `bson:"group"`
	Teacher      string             `bson:"teacher"`
	Room         string             `bson:"room"`
	DayIndex     int                `bson:"dayIndex"`
	WeekIndex    int                `bson:"weekIndex"`
	ParsedInfo   map[string]any     `bson:"parsedInfo"`
}

type Lesson struct {
	Id           primitive.ObjectID `bson:"_id"`
	StudyPlaceId int                `bson:"studyPlaceId"`
	Type         string             `bson:"type"`
	EndDate      time.Time          `bson:"endDate"`
	StartDate    time.Time          `bson:"startDate"`
	Subject      string             `bson:"subject"`
	Group        string             `bson:"group"`
	Teacher      string             `bson:"teacher"`
	Room         string             `bson:"room"`
	Marks        []Mark             `bson:"marks"`
	Title        string             `bson:"title"`
	Homework     string             `bson:"homework"`
	Description  string             `bson:"description"`
	ParsedInfo   map[string]any     `bson:"parsedInfo"`
}

type Mark struct {
	Id           primitive.ObjectID `bson:"_id"`
	Mark         string             `bson:"mark"`
	UserId       primitive.ObjectID `bson:"userId"`
	LessonId     primitive.ObjectID `bson:"lessonId"`
	StudyPlaceId int                `bson:"studyPlaceId"`
	ParsedInfo   map[string]any     `bson:"parsedInfo"`
}
