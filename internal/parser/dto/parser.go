package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/parser/entities"
	"time"
)

type LessonDTO struct {
	Shift      entities.Shift
	Type       string
	Subject    string
	Group      string
	Teacher    string
	Room       string
	ParsedInfo map[string]any
}

type GeneralLessonDTO struct {
	Shift      entities.Shift
	Subject    string
	Group      string
	Teacher    string
	Room       string
	WeekIndex  int
	ParsedInfo map[string]any
}

type MarkDTO struct {
	Mark       string
	UserId     primitive.ObjectID
	LessonDate time.Time
	Subject    string
	Group      string
	ParsedInfo map[string]any
}
