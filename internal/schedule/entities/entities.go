package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/journal/entities"
	"time"
)

type Schedule struct {
	Info    Info     `json:"info" bson:"info"`
	Lessons []Lesson `json:"lessons" bson:"lessons"`
}

type GeneralSchedule struct {
	Info           GeneralInfo     `json:"info" bson:"info"`
	GeneralLessons []GeneralLesson `json:"lessons" bson:"lessons"`
}

type DeleteLessonID struct {
	ID primitive.ObjectID `apps:"trackable,collection=Lessons"`
}

type Lesson struct {
	Id               primitive.ObjectID `json:"id" bson:"_id" apps:"trackable,collection=Lessons"`
	StudyPlaceId     primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
	PrimaryColor     string             `json:"primaryColor" bson:"primaryColor"`
	JournalCellColor string             `json:"journalCellColor" bson:"journalCellColor"`
	SecondaryColor   string             `json:"secondaryColor" bson:"secondaryColor"`
	Type             string             `json:"type" bson:"type"`
	EndDate          time.Time          `json:"endDate" bson:"endDate"`
	StartDate        time.Time          `json:"startDate" bson:"startDate"`
	LessonIndex      int                `json:"lessonIndex" bson:"lessonIndex"`
	Marks            []entities.Mark    `json:"marks" bson:"marks"`
	Absences         []entities.Absence `json:"absences" bson:"absences"`
	Subject          string             `json:"subject" bson:"subject"`
	Group            string             `json:"group" bson:"group"`
	Teacher          string             `json:"teacher" bson:"teacher"`
	Room             string             `json:"room" bson:"room"`
	SubjectID        primitive.ObjectID `json:"subjectID" bson:"subjectID"`
	GroupID          primitive.ObjectID `json:"groupID" bson:"groupID"`
	TeacherID        primitive.ObjectID `json:"teacherID" bson:"teacherID"`
	RoomID           primitive.ObjectID `json:"roomID" bson:"roomID"`
	Title            string             `json:"title" bson:"title"`
	Homework         string             `json:"homework" bson:"homework"`
	Description      string             `json:"description" bson:"description"`
	IsGeneral        bool               `json:"isGeneral" bson:"isGeneral"`
	Status           string             `json:"status" bson:"status"`
}

type GeneralLesson struct {
	Id               primitive.ObjectID `json:"id" bson:"_id"`
	StudyPlaceId     primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
	PrimaryColor     string             `json:"primaryColor" bson:"primaryColor"`
	SecondaryColor   string             `json:"secondaryColor" bson:"secondaryColor"`
	Subject          string             `json:"subject" bson:"subject"`
	Group            string             `json:"group" bson:"group"`
	Teacher          string             `json:"teacher" bson:"teacher"`
	Room             string             `json:"room" bson:"room"`
	SubjectID        primitive.ObjectID `json:"subjectID" bson:"subjectID"`
	GroupID          primitive.ObjectID `json:"groupID" bson:"groupID"`
	TeacherID        primitive.ObjectID `json:"teacherID" bson:"teacherID"`
	RoomID           primitive.ObjectID `json:"roomID" bson:"roomID"`
	Type             string             `json:"type" bson:"type"`
	LessonIndex      int                `json:"lessonIndex" bson:"lessonIndex"`
	DayIndex         int                `json:"dayIndex" bson:"dayIndex"`
	WeekIndex        int                `json:"weekIndex" bson:"weekIndex"`
	StartTimeMinutes int                `json:"startTimeMinutes" bson:"startTimeMinutes"`
	EndTimeMinutes   int                `json:"endTimeMinutes" bson:"endTimeMinutes"`
}

type Info struct {
	StudyPlaceInfo StudyPlaceInfo `json:"studyPlaceInfo" bson:"studyPlaceInfo"`
	Type           string         `json:"type" bson:"type"`
	TypeName       string         `json:"typeName" bson:"typeName"`
	StartDate      time.Time      `json:"startDate" bson:"startDate"`
	EndDate        time.Time      `json:"endDate" bson:"endDate"`
	Date           time.Time      `json:"date" bson:"date"`
}

type GeneralInfo struct {
	StudyPlaceInfo StudyPlaceInfo `json:"studyPlaceInfo" bson:"studyPlaceInfo"`
	Type           string         `json:"type" bson:"type"`
	TypeName       string         `json:"typeName" bson:"typeName"`
	Date           time.Time      `json:"date" bson:"date"`
}

type StudyPlaceInfo struct {
	Id    primitive.ObjectID `json:"id" bson:"_id"`
	Title string             `json:"title" bson:"title"`
}

type Types struct {
	Groups   []TypeEntry `json:"groups" bson:"groups"`
	Teachers []TypeEntry `json:"teachers" bson:"teachers"`
	Subjects []TypeEntry `json:"subjects" bson:"subjects"`
	Rooms    []TypeEntry `json:"rooms" bson:"rooms"`
}

type TypeEntry struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Title string             `json:"title" bson:"title"`
}

type ScheduleInfoEntry struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Date         time.Time          `json:"date" bson:"date"`
	Status       string             `json:"status" bson:"status"`
	StudyPlaceId primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
}
