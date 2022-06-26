package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ParseJournalUser struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	ParserAppName  string             `bson:"parserAppName" json:"parserAppName"`
	Login          string             `bson:"login" json:"login"`
	Password       string             `bson:"password" json:"password"`
	AdditionInfo   map[string]string  `bson:"additionInfo" json:"additionInfo"`
	LastParsedDate time.Time          `bson:"lastParsedDate" json:"lastParsedDate"`
}
