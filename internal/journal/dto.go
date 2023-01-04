package journal

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AddMarkDTO struct {
	Mark      string             `json:"mark"`
	StudentID primitive.ObjectID `json:"studentID" binding:"required"`
	LessonID  primitive.ObjectID `json:"lessonId" binding:"required"`
}

type UpdateMarkDTO struct {
	ID primitive.ObjectID `json:"id" binding:"required"`
	AddMarkDTO
}

type AddAbsencesDTO struct {
	Time      *int               `json:"time"`
	StudentID primitive.ObjectID `json:"studentID" binding:"required"`
	LessonID  primitive.ObjectID `json:"lessonID" binding:"required"`
}

type UpdateAbsencesDTO struct {
	ID primitive.ObjectID `json:"id" binding:"required"`
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
