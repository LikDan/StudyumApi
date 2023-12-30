package dtos

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AddMarkDTO struct {
	MarkID    primitive.ObjectID `json:"markID"`
	StudentID primitive.ObjectID `json:"studentID"`
	LessonID  primitive.ObjectID `json:"lessonID"`
}

type UpdateMarkDTO struct {
	ID primitive.ObjectID `json:"id"`
	AddMarkDTO
}

type AddAbsencesDTO struct {
	Time      *int               `json:"time"`
	StudentID primitive.ObjectID `json:"studentID"`
	LessonID  primitive.ObjectID `json:"lessonID"`
}

type UpdateAbsencesDTO struct {
	ID primitive.ObjectID `json:"id"`
	AddAbsencesDTO
}

type MarksReport struct {
	LessonType string     `json:"lessonType" bson:"lessonType"`
	Mark       string     `json:"mark" bson:"mark"`
	StartDate  *time.Time `json:"startDate" bson:"startDate"`
	EndDate    *time.Time `json:"endDate" bson:"endDate"`
	NotExists  bool       `json:"notExists" bson:"notExists"`
}

type AbsencesReport struct {
	StartDate *time.Time `json:"startDate" bson:"startDate"`
	EndDate   *time.Time `json:"endDate" bson:"endDate"`
}
