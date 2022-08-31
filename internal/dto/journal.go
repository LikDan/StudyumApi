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
	Id        primitive.ObjectID `json:"id" binding:"required"`
	Mark      string             `json:"mark"`
	StudentID primitive.ObjectID `json:"studentID" binding:"required"`
	LessonId  primitive.ObjectID `json:"lessonId" binding:"required"`
}
