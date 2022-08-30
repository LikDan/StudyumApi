package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/entities"
	"time"
)

type LessonDTO struct {
	Shift          entities.Shift
	PrimaryColor   string
	SecondaryColor string
	Subject        string
	Group          string
	Teacher        string
	Room           string
	ParsedInfo     entities.ParsedInfoType
}

type GeneralLessonDTO struct {
	Shift      entities.Shift
	Subject    string
	Group      string
	Teacher    string
	Room       string
	WeekIndex  int
	ParsedInfo entities.ParsedInfoType
}

type MarkDTO struct {
	Mark       string
	StudentID  primitive.ObjectID
	LessonDate time.Time
	Subject    string
	Group      string
	ParsedInfo entities.ParsedInfoType
}
