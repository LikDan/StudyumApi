package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AcceptUser struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	Name     string             `json:"name" bson:"name" encryption:""`
	Role     string             `json:"role" bson:"role"`
	RoleName string             `json:"roleName" bson:"roleName"`
}

type SignUpCode struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	Code         string             `json:"code" bson:"code"`
	Name         string             `json:"name" bson:"name" encryption:""`
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
	Role         string             `json:"role" bson:"role"`
	RoleName     string             `json:"roleName" bson:"roleName"`
	Password     string             `json:"-" bson:"defaultPassword"`
}
