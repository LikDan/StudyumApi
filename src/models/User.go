package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id            primitive.ObjectID `json:"id" bson:"_id"`
	Token         string             `json:"-" bson:"token"`
	Password      string             `json:"password" bson:"password"`
	Email         string             `json:"email" bson:"email"`
	VerifiedEmail bool               `json:"verifiedEmail" bson:"verifiedEmail"`
	Login         string             `json:"login" bson:"login"`
	Name          string             `json:"name" bson:"name"`
	PictureUrl    string             `json:"picture" bson:"picture"`
	Type          string             `json:"type" bson:"type"`
	TypeName      string             `json:"typeName" bson:"typeName"`
	StudyPlaceId  int                `json:"studyPlaceId" bson:"studyPlaceId"`
	Permissions   []string           `json:"permissions" bson:"permissions"`
	Accepted      bool               `json:"accepted" bson:"accepted"`
	Blocked       bool               `json:"blocked" bson:"blocked"`
}
