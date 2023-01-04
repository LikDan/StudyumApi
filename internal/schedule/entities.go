package schedule

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/general"
	"studyum/internal/journal"
	parser "studyum/internal/parser/entities"
	"time"
)

type Schedule struct {
	Info    Info     `json:"info" bson:"info"`
	Lessons []Lesson `json:"lessons" bson:"lessons"`
}

type Lesson struct {
	Id               primitive.ObjectID    `json:"id" bson:"_id"`
	StudyPlaceId     primitive.ObjectID    `json:"studyPlaceId" bson:"studyPlaceId"`
	PrimaryColor     string                `json:"primaryColor" bson:"primaryColor"`
	JournalCellColor string                `json:"journalCellColor" bson:"journalCellColor"`
	SecondaryColor   string                `json:"secondaryColor" bson:"secondaryColor"`
	Type             string                `json:"type" bson:"type"`
	EndDate          time.Time             `json:"endDate" bson:"endDate"`
	StartDate        time.Time             `json:"startDate" bson:"startDate"`
	Marks            []journal.Mark        `json:"marks" bson:"marks"`
	Absences         []journal.Absence     `json:"absences" bson:"absences"`
	Subject          string                `json:"subject" bson:"subject"`
	Group            string                `json:"group" bson:"group"`
	Teacher          string                `json:"teacher" bson:"teacher"`
	Room             string                `json:"room" bson:"room"`
	Title            string                `json:"title" bson:"title"`
	Homework         string                `json:"homework" bson:"homework"`
	Description      string                `json:"description" bson:"description"`
	IsGeneral        bool                  `json:"isGeneral" bson:"isGeneral"`
	ParsedInfo       parser.ParsedInfoType `json:"-" bson:"parsedInfo"`
}

type GeneralLesson struct {
	Id             primitive.ObjectID    `json:"id" bson:"_id"`
	StudyPlaceId   primitive.ObjectID    `json:"studyPlaceId" bson:"studyPlaceId"`
	PrimaryColor   string                `json:"primaryColor" bson:"primaryColor"`
	SecondaryColor string                `json:"secondaryColor" bson:"secondaryColor"`
	EndTime        string                `json:"endTime" bson:"endTime"`
	StartTime      string                `json:"startTime" bson:"startTime"`
	Subject        string                `json:"subject" bson:"subject"`
	Group          string                `json:"group" bson:"group"`
	Teacher        string                `json:"teacher" bson:"teacher"`
	Room           string                `json:"room" bson:"room"`
	Type           string                `json:"type" bson:"type"`
	DayIndex       int                   `json:"dayIndex" bson:"dayIndex"`
	WeekIndex      int                   `json:"weekIndex" bson:"weekIndex"`
	ParsedInfo     parser.ParsedInfoType `json:"-" bson:"parsedInfo"`
}

type Info struct {
	Type          string             `json:"type" bson:"type"`
	TypeName      string             `json:"typeName" bson:"typeName"`
	StudyPlace    general.StudyPlace `json:"studyPlace" bson:"studyPlace"`
	StartWeekDate time.Time          `json:"startWeekDate" bson:"startWeekDate"`
	Date          time.Time          `json:"date" bson:"date"`
}

type Types struct {
	Groups   []string `json:"groups" bson:"groups"`
	Teachers []string `json:"teachers" bson:"teachers"`
	Subjects []string `json:"subjects" bson:"subjects"`
	Rooms    []string `json:"rooms" bson:"rooms"`
}
