package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type ScheduleTypeInfo struct {
	Id            primitive.ObjectID `bson:"_id" json:"id"`
	ParserAppName string             `bson:"parserAppName" json:"parserAppName"`
	Group         string             `bson:"group" json:"group"`
	Url           string             `bson:"url" json:"url"`
}
