package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Mark struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	Mark         string             `json:"mark" bson:"mark"`
	UserId       primitive.ObjectID `json:"userId" bson:"userId"`
	LessonId     primitive.ObjectID `json:"lessonId" bson:"lessonId"`
	StudyPlaceId int                `json:"studyPlaceId" bson:"studyPlaceId"`
}
