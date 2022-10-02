package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AddMarkDTO struct {
	Mark      string             `json:"mark"`
	StudentID primitive.ObjectID `json:"studentID" binding:"required"`
	LessonId  primitive.ObjectID `json:"lessonId" binding:"required"`
}

type UpdateMarkDTO struct {
	Id primitive.ObjectID `json:"id" binding:"required"`
	AddMarkDTO
}

type AddAbsencesDTO struct {
	Time      *int               `json:"time"`
	StudentID primitive.ObjectID `json:"studentID" binding:"required"`
	LessonID  primitive.ObjectID `json:"lessonID" binding:"required"`
}

type UpdateAbsencesDTO struct {
	Id primitive.ObjectID `json:"id" binding:"required"`
	AddAbsencesDTO
}
