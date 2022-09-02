package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AddLessonDTO struct {
	PrimaryColor   string    `json:"primaryColor" binding:"hexcolor|eq=transparent"`
	SecondaryColor string    `json:"secondaryColor" binding:"hexcolor|eq=transparent"`
	EndDate        time.Time `json:"endDate"`
	StartDate      time.Time `json:"startDate"`
	Subject        string    `json:"subject" binding:"required"`
	Group          string    `json:"group" binding:"required"`
	Teacher        string    `json:"teacher" binding:"required"`
	Room           string    `json:"room" binding:"required"`
}

type UpdateLessonDTO struct {
	Id          primitive.ObjectID `json:"id" binding:"required"`
	Subject     string             `json:"subject" binding:"required"`
	Group       string             `json:"group" binding:"required"`
	Teacher     string             `json:"teacher" binding:"required"`
	Room        string             `json:"room" binding:"required"`
	Title       string             `json:"title"`
	Homework    string             `json:"homework"`
	Description string             `json:"description"`
}
