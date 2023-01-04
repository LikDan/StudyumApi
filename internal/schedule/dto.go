package schedule

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AddGeneralLessonDTO struct {
	PrimaryColor   string `json:"primaryColor" binding:"hexcolor|eq=transparent"`
	SecondaryColor string `json:"secondaryColor" binding:"hexcolor|eq=transparent"`
	DayIndex       int    `json:"dayIndex"`
	WeekIndex      int    `json:"weekIndex"`
	StartTime      string `json:"startTime" binding:"required"`
	EndTime        string `json:"endTime" binding:"required"`
	Subject        string `json:"subject" binding:"required"`
	Teacher        string `json:"teacher" binding:"required"`
	Group          string `json:"group" binding:"required"`
	Room           string `json:"room" binding:"required"`
}

type AddLessonDTO struct {
	PrimaryColor   string    `json:"primaryColor" binding:"hexcolor|eq=transparent"`
	SecondaryColor string    `json:"secondaryColor" binding:"hexcolor|eq=transparent"`
	EndDate        time.Time `json:"endDate"`
	StartDate      time.Time `json:"startDate"`
	Type           string    `json:"type" binding:"required"`
	Subject        string    `json:"subject"`
	Group          string    `json:"group" binding:"required"`
	Teacher        string    `json:"teacher" binding:"required"`
	Room           string    `json:"room" binding:"required"`
}

type UpdateLessonDTO struct {
	AddLessonDTO
	Id          primitive.ObjectID `json:"id" binding:"required"`
	Title       string             `json:"title"`
	Homework    string             `json:"homework"`
	Description string             `json:"description"`
}
