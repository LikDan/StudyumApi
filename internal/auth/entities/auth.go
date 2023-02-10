package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id            primitive.ObjectID `json:"id" bson:"_id"`
	Password      string             `json:"-" bson:"password"`
	Email         string             `json:"email" bson:"email"`
	VerifiedEmail bool               `json:"verifiedEmail" bson:"verifiedEmail"`
	FirebaseToken string             `json:"-" bson:"firebaseToken" encryption:""`
	Login         string             `json:"login" bson:"login"`
	Name          string             `json:"name" bson:"name" encryption:""`
	PictureUrl    string             `json:"picture" bson:"picture" encryption:""`
	Type          string             `json:"type" bson:"type"`
	TypeName      string             `json:"typeName" bson:"typename"`
	TuitionGroup  string             `json:"tuitionGroup" bson:"tuitionGroup"`
	StudyPlaceID  primitive.ObjectID `json:"studyPlaceId" bson:"studyPlaceID"`
	Permissions   []string           `json:"permissions" bson:"permissions"`
	Accepted      bool               `json:"accepted" bson:"accepted"`
	Blocked       bool               `json:"blocked" bson:"blocked"`
}
