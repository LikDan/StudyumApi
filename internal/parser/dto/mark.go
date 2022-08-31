package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/entities"
)

type MarkDTO struct {
	Id           primitive.ObjectID
	Mark         string
	StudentID    primitive.ObjectID
	LessonId     primitive.ObjectID
	StudyPlaceId int
	ParsedInfo   entities.ParsedInfoType
}
