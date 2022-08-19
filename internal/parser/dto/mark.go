package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type Mark struct {
	Id           primitive.ObjectID
	Mark         string
	UserId       primitive.ObjectID
	LessonId     primitive.ObjectID
	StudyPlaceId int
	ParsedInfo   map[string]any
}
