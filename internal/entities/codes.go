package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type SignUpCode struct {
	Id           primitive.ObjectID `json:"_id"`
	Code         string             `json:"code"`
	Name         string             `json:"name" encryption:""`
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID"`
	Type         string             `json:"type"`
	Typename     string             `json:"typename"`
}
