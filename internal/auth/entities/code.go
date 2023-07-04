package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserCodeData struct {
	Id              primitive.ObjectID `json:"id" bson:"_id"`
	Code            string             `json:"code" bson:"code"`
	Name            string             `json:"name" bson:"name" encryption:""`
	StudyPlaceID    primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
	Role            string             `json:"role" bson:"role"`
	RoleName        string             `json:"roleName" bson:"roleName"`
	TuitionGroup    string             `json:"tuitionGroup" bson:"tuitionGroup"`
	Permissions     []string           `json:"permissions" bson:"permissions"`
	DefaultPassword string             `json:"defaultPassword" bson:"defaultPassword"`
}
