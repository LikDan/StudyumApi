package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Lesson struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	StudyPlaceId int                `json:"studyPlaceId" bson:"studyPlaceId"`
	Type         string             `json:"type" bson:"type"`
	EndDate      time.Time          `json:"endDate" bson:"endDate"`
	StartDate    time.Time          `json:"startDate" bson:"startDate"`
	Subject      string             `json:"subject" bson:"subject"`
	Group        string             `json:"group" bson:"group"`
	Teacher      string             `json:"teacher" bson:"teacher"`
	Room         string             `json:"room" bson:"room"`
	Marks        []Mark             `json:"marks" bson:"marks"`
	Title        string             `json:"title" bson:"title"`
	Homework     string             `json:"homework" bson:"homework"`
	Description  string             `json:"description" bson:"description"`
}
