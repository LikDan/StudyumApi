package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/entities"
	"time"
)

type Lesson struct {
	Id           primitive.ObjectID
	StudyPlaceId int
	Type         string
	EndDate      time.Time
	StartDate    time.Time
	Subject      string
	Group        string
	Teacher      string
	Room         string
	ParsedInfo   entities.ParsedInfoType
}
