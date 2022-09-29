package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type StudyPlace struct {
	Id                primitive.ObjectID `json:"id" bson:"_id"`
	WeeksCount        int                `json:"weeksCount" bson:"weeksCount"`
	Name              string             `json:"name" bson:"name"`
	PrimaryColorSet   []string           `json:"primaryColorSet" bson:"primaryColorSet"`
	SecondaryColorSet []string           `json:"secondaryColorSet" bson:"secondaryColorSet"`
	LessonTypes       []LessonType       `json:"lessonTypes" bson:"lessonTypes"`
	Restricted        bool               `json:"restricted" bson:"restricted"`
	AdminID           primitive.ObjectID `json:"adminID" bson:"adminID"`
}

type MarkType struct {
	Mark       string `bson:"mark" json:"mark"`
	Standalone bool   `bson:"standalone" json:"standalone"`
}

type LessonType struct {
	Type  string     `bson:"type" json:"type"`
	Marks []MarkType `bson:"marks" json:"marks"`
}
