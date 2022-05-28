package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type GeneralLesson struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	StudyPlaceId int                `json:"studyPlaceId" bson:"studyPlaceId"`
	EndTime      string             `json:"endTime" bson:"endTime"`
	StartTime    string             `json:"startTime" bson:"startTime"`
	Subject      string             `json:"subject" bson:"subject"`
	Group        string             `json:"group" bson:"group"`
	Teacher      string             `json:"teacher" bson:"teacher"`
	Room         string             `json:"room" bson:"room"`
	DayIndex     int                `json:"dayIndex" bson:"dayIndex"`
	WeekIndex    int                `json:"weekIndex" bson:"weekIndex"`
}
