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
	Id       string   `json:"id" bson:"_id"`
	Title    string   `json:"title" bson:"title" encryption:""`
	UserType string   `json:"userType" bson:"userType"`
	Lessons  []Lesson `json:"lessons" bson:"lessons"`
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
	LessonId     primitive.ObjectID      `json:"lessonId" bson:"lessonId"`
	StudyPlaceId primitive.ObjectID      `json:"studyPlaceId" bson:"studyPlaceId"`
	ParsedInfo   entities.ParsedInfoType `json:"-" bson:"parsedInfo"`
}
