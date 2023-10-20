package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id             primitive.ObjectID  `json:"id" bson:"_id"`
	Password       string              `json:"-" bson:"password"`
	Email          string              `json:"email" bson:"email"`
	VerifiedEmail  bool                `json:"verifiedEmail" bson:"verifiedEmail"`
	FirebaseToken  string              `json:"-" bson:"firebaseToken" encryption:""`
	Login          string              `json:"login" bson:"login"`
	PictureUrl     string              `json:"picture" bson:"picture" encryption:""`
	StudyPlaceInfo *UserStudyPlaceInfo `json:"studyPlaceInfo" bson:"studyPlaceInfo" encryption:""`
}

type UserStudyPlaceInfo struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name" encryption:""`
	Role         string             `json:"role" bson:"role"`
	RoleName     string             `json:"roleName" bson:"roleName"`
	TuitionGroup string             `json:"tuitionGroup" bson:"tuitionGroup"`
	Permissions  []string           `json:"permissions" bson:"permissions"`
	Accepted     bool               `json:"accepted" bson:"accepted"`
}
