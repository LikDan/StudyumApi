package user

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OAuth2CallbackUser struct {
	Id            string `json:"id" bson:"_id"`
	Email         string `json:"email" bson:"email"`
	VerifiedEmail bool   `json:"verified_email" bson:"verifiedEmail"`
	Name          string `json:"name" bson:"login"`
	PictureUrl    string `json:"picture" bson:"picture"`
}

type AcceptUser struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	Name     string             `json:"name" bson:"name" encryption:""`
	Type     string             `json:"type" bson:"type"`
	Typename string             `json:"typename" bson:"typename"`
}

type SignUpCode struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	Code         string             `json:"code" bson:"code"`
	Name         string             `json:"name" bson:"name" encryption:""`
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
	Type         string             `json:"type" bson:"type"`
	Typename     string             `json:"typename" bson:"typename"`
}
