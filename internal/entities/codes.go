package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type SignUpCode struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	Code         string             `json:"code" bson:"code"`
	Name         string             `json:"name" bson:"name" encryption:""`
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
	Type         string             `json:"type" bson:"type"`
	Typename     string             `json:"typename" bson:"typename"`
}
