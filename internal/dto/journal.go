package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AddMarkDTO struct {
	Mark     string             `json:"mark"`
	UserId   primitive.ObjectID `json:"userId"`
	LessonId primitive.ObjectID `json:"lessonId"`
}

type UpdateMarkDTO struct {
	Id       primitive.ObjectID `json:"id"`
	Mark     string             `json:"mark"`
	UserId   primitive.ObjectID `json:"userId"`
	LessonId primitive.ObjectID `json:"lessonId"`
}
