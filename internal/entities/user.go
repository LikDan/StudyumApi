package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id            primitive.ObjectID `json:"id" bson:"_id"`
	Token         string             `json:"-" bson:"token"`
	Password      string             `json:"password" bson:"password"`
	Email         string             `json:"email" bson:"email"`
	FirebaseToken string             `json:"-" bson:"firebaseToken"`
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

type OAuth2CallbackUser struct {
	Id            string `json:"id" bson:"_id"`
	Email         string `json:"email" bson:"email"`
	VerifiedEmail bool   `json:"verified_email" bson:"verifiedEmail"`
	Name          string `json:"name" bson:"login"`
	PictureUrl    string `json:"picture" bson:"picture"`
}
