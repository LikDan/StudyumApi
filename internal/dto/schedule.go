package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AddLessonDTO struct {
	Type      string    `json:"type"`
	EndDate   time.Time `json:"endDate"`
	StartDate time.Time `json:"startDate"`
	Subject   string    `json:"subject"`
	Group     string    `json:"group"`
	Teacher   string    `json:"teacher"`
	Room      string    `json:"room"`
}

type UpdateLessonDTO struct {
	Id          primitive.ObjectID `json:"id"`
	Subject     string             `json:"subject"`
	Group       string             `json:"group"`
	Teacher     string             `json:"teacher"`
	Room        string             `json:"room"`
	Title       string             `json:"title"`
	Homework    string             `json:"homework"`
	Description string             `json:"description"`
}
