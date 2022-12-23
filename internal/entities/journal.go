package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/entities"
)

type Journal struct {
	Info  JournalInfo  `json:"info" bson:"info"`
	Rows  []JournalRow `json:"rows" bson:"rows"`
	Dates []Lesson     `json:"dates" bson:"dates"`
}

type JournalInfo struct {
	Editable   bool       `json:"editable" bson:"editable"`
	StudyPlace StudyPlace `json:"studyPlace" bson:"studyPlace"`
	Group      string     `json:"group" bson:"group"`
	Teacher    string     `json:"teacher" bson:"teacher"`
	Subject    string     `json:"subject" bson:"subject"`
}

type JournalRow struct {
	Id                 string       `json:"id" bson:"_id"`
	Title              string       `json:"title" bson:"title"`
	Lessons            []*Lesson    `json:"lessons" bson:"lessons"`
	NumericMarksSum    int          `json:"numericMarksSum" bson:"numericMarksSum"`
	NumericMarksLength int          `json:"numericMarksAmount" bson:"numericMarksAmount"`
	AbsencesAmount     int          `json:"absencesAmount" bson:"absencesAmount"`
	AbsencesTime       int          `json:"absencesTime" bson:"absencesTime"`
	MarksAmount        []MarkAmount `json:"marksAmount" bson:"marksAmount"`
	Color              string       `json:"color" bson:"color"`
}

type JournalAvailableOption struct {
	Teacher  string `json:"teacher"`
	Subject  string `json:"subject"`
	Group    string `json:"group"`
	Editable bool   `json:"editable"`
}

type Mark struct {
	Id           primitive.ObjectID      `json:"id" bson:"_id"`
	Mark         string                  `json:"mark" bson:"mark"`
	StudentID    primitive.ObjectID      `json:"studentID" bson:"studentID"`
	LessonID     primitive.ObjectID      `json:"lessonID" bson:"lessonID"`
	StudyPlaceID primitive.ObjectID      `json:"studyPlaceID" bson:"studyPlaceID"`
	ParsedInfo   entities.ParsedInfoType `json:"-" bson:"parsedInfo"`
}

type Absence struct {
	Id           primitive.ObjectID      `json:"id" bson:"_id"`
	Time         *int                    `json:"time" bson:"time"`
	StudentID    primitive.ObjectID      `json:"studentID" bson:"studentID"`
	LessonID     primitive.ObjectID      `json:"lessonID" bson:"lessonID"`
	StudyPlaceID primitive.ObjectID      `json:"studyPlaceID" bson:"studyPlaceID"`
	ParsedInfo   entities.ParsedInfoType `json:"-" bson:"parsedInfo"`
}

type MarkAmount struct {
	Mark   string `json:"mark" bson:"mark"`
	Amount int    `json:"amount" bson:"amount"`
}
