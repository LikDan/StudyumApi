package dtos

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AddMarkDTO struct {
	Mark      string             `json:"mark"`
	StudentID primitive.ObjectID `json:"studentID" binding:"req"`
	LessonID  primitive.ObjectID `json:"lessonId" binding:"req"`
}

type UpdateMarkDTO struct {
	ID primitive.ObjectID `json:"id" binding:"req"`
	AddMarkDTO
}

type AddAbsencesDTO struct {
	Time      *int               `json:"time"`
	StudentID primitive.ObjectID `json:"studentID" binding:"req"`
	LessonID  primitive.ObjectID `json:"lessonID" binding:"req"`
}

type UpdateAbsencesDTO struct {
	ID primitive.ObjectID `json:"id" binding:"req"`
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
