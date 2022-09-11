package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type StudyPlace struct {
	Id                primitive.ObjectID `json:"id" bson:"_id"`
	WeeksCount        int                `json:"weeksCount" bson:"weeksCount"`
	Name              string             `json:"name" bson:"name"`
	PrimaryColorSet   []string           `json:"primaryColorSet" bson:"primaryColorSet"`
	SecondaryColorSet []string           `json:"secondaryColorSet" bson:"secondaryColorSet"`
	Restricted        bool               `json:"restricted" bson:"restricted"`
	AdminID           primitive.ObjectID `json:"adminID" bson:"adminID"`
}
