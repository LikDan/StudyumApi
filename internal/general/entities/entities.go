package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type StudyPlace struct {
	Id                primitive.ObjectID `json:"id" bson:"_id"`
	WeeksCount        int                `json:"weeksCount" bson:"weeksCount"`
	Name              string             `json:"name" bson:"name"`
	Description       string             `json:"description" bson:"description"`
	PictureUrl        string             `json:"picture" bson:"picture"`
	BannerUrl         string             `json:"banner" bson:"banner"`
	Address           string             `json:"address" bson:"address"`
	Phone             string             `json:"phone" bson:"phone"`
	PrimaryColorSet   []string           `json:"primaryColorSet" bson:"primaryColorSet"`
	SecondaryColorSet []string           `json:"secondaryColorSet" bson:"secondaryColorSet"`
	JournalColors     JournalColors      `json:"journalColors" bson:"journalColors"`
	LessonTypes       []LessonType       `json:"lessonTypes" bson:"lessonTypes"`
	Restricted        bool               `json:"restricted" bson:"restricted"`
	AdminID           primitive.ObjectID `json:"adminID" bson:"adminID"`
	AbsenceMark       string             `json:"absenceMark" bson:"absenceMark"`
	IsSchedulePrivate bool               `json:"IsSchedulePrivate" bson:"IsSchedulePrivate"`
}

type MarkType struct {
	Mark        string        `bson:"mark" json:"mark"`
	WorkOutTime time.Duration `bson:"workOutTime" json:"workOutTime"`
}

type LessonType struct {
	Type               string        `bson:"type" json:"type"`
	AbsenceWorkOutTime time.Duration `bson:"absenceWorkOutTime" json:"absenceWorkOutTime"`
	Marks              []MarkType    `bson:"marks" json:"marks"`
	AssignedColor      string        `bson:"assignedColor" json:"assignedColor"`
	StandaloneMarks    []MarkType    `bson:"standaloneMarks" json:"standaloneMarks"`
}

type JournalColors struct {
	General string `json:"general"`
	Warning string `json:"warning"`
	Danger  string `json:"danger"`
}
