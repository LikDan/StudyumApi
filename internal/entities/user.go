package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id            primitive.ObjectID `json:"id" bson:"_id"`
	Password      string             `json:"password" bson:"password"`
	Email         string             `json:"email" bson:"email"`
	VerifiedEmail bool               `json:"verifiedEmail" bson:"verifiedEmail"`
	FirebaseToken string             `json:"-" bson:"firebaseToken" encryption:""`
	Login         string             `json:"login" bson:"login" encryption:""`
	Name          string             `json:"name" bson:"name" encryption:""`
	PictureUrl    string             `json:"picture" bson:"picture" encryption:""`
	Type          string             `json:"type" bson:"type"`
	TypeName      string             `json:"typeName" bson:"typeName"`
	StudyPlaceId  primitive.ObjectID `json:"studyPlaceId" bson:"studyPlaceId"`
	Permissions   []string           `json:"permissions" bson:"permissions"`
	Accepted      bool               `json:"accepted" bson:"accepted"`
	Blocked       bool               `json:"blocked" bson:"blocked"`
	Sessions      []Session          `json:"sessions" bson:"sessions"`
}

type OAuth2CallbackUser struct {
	Id            string `json:"id" bson:"_id"`
	Email         string `json:"email" bson:"email"`
	VerifiedEmail bool   `json:"verified_email" bson:"verifiedEmail"`
	Name          string `json:"name" bson:"login"`
	PictureUrl    string `json:"picture" bson:"picture"`
}
