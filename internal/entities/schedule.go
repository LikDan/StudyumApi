package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/entities"
	"time"
)

type Schedule struct {
	Info    ScheduleInfo `json:"info" bson:"info"`
	Lessons []Lesson     `json:"lessons" bson:"lessons"`
}

type Lesson struct {
	Id             primitive.ObjectID      `json:"id" bson:"_id"`
	StudyPlaceId   int                     `json:"studyPlaceId" bson:"studyPlaceId"`
	PrimaryColor   string                  `json:"primaryColor" bson:"primaryColor"`
	SecondaryColor string                  `json:"secondaryColor" bson:"secondaryColor"`
	EndDate        time.Time               `json:"endDate" bson:"endDate"`
	StartDate      time.Time               `json:"startDate" bson:"startDate"`
	Subject        string                  `json:"subject" bson:"subject"`
	Group          string                  `json:"group" bson:"group"`
	Teacher        string                  `json:"teacher" bson:"teacher"`
	Room           string                  `json:"room" bson:"room"`
	Marks          []Mark                  `json:"marks" bson:"marks"`
	Title          string                  `json:"title" bson:"title"`
	Homework       string                  `json:"homework" bson:"homework"`
	Description    string                  `json:"description" bson:"description"`
	ParsedInfo     entities.ParsedInfoType `json:"-" bson:"parsedInfo"`
}

type GeneralLesson struct {
	Id             primitive.ObjectID      `json:"id" bson:"_id"`
	StudyPlaceId   int                     `json:"studyPlaceId" bson:"studyPlaceId"`
	PrimaryColor   string                  `json:"primaryColor" bson:"primaryColor"`
	SecondaryColor string                  `json:"secondaryColor" bson:"secondaryColor"`
	EndTime        string                  `json:"endTime" bson:"endTime"`
	StartTime      string                  `json:"startTime" bson:"startTime"`
	Subject        string                  `json:"subject" bson:"subject"`
	Group          string                  `json:"group" bson:"group"`
	Teacher        string                  `json:"teacher" bson:"teacher"`
	Room           string                  `json:"room" bson:"room"`
	DayIndex       int                     `json:"dayIndex" bson:"dayIndex"`
	WeekIndex      int                     `json:"weekIndex" bson:"weekIndex"`
	ParsedInfo     entities.ParsedInfoType `json:"-" bson:"parsedInfo"`
}

type ScheduleInfo struct {
	Type          string     `json:"type" bson:"type"`
	TypeName      string     `json:"typeName" bson:"typeName"`
	StudyPlace    StudyPlace `json:"studyPlace" bson:"studyPlace"`
	StartWeekDate time.Time  `json:"startWeekDate" bson:"startWeekDate"`
	Date          time.Time  `json:"date" bson:"date"`
}

type Types struct {
	Groups   []string `json:"groups" bson:"groups"`
	Teachers []string `json:"teachers" bson:"teachers"`
	Subjects []string `json:"subjects" bson:"subjects"`
	Rooms    []string `json:"rooms" bson:"rooms"`
}
