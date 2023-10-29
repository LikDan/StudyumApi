package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AddGeneralLessonDTO struct {
	PrimaryColor   string `json:"primaryColor" binding:"hexcolor|eq=transparent"`
	SecondaryColor string `json:"secondaryColor" binding:"hexcolor|eq=transparent"`
	LessonIndex    int    `json:"lessonIndex"`
	DayIndex       int    `json:"dayIndex"`
	WeekIndex      int    `json:"weekIndex"`
	StartTime      string `json:"startTime" binding:"req"`
	EndTime        string `json:"endTime" binding:"req"`
	Subject        string `json:"subject" binding:"req"`
	Teacher        string `json:"teacher" binding:"req"`
	Group          string `json:"group" binding:"req"`
	Room           string `json:"room" binding:"req"`
}

type AddLessonDTO struct {
	PrimaryColor   string             `json:"primaryColor"`
	SecondaryColor string             `json:"secondaryColor"`
	EndDate        time.Time          `json:"endDate"`
	StartDate      time.Time          `json:"startDate"`
	LessonIndex    int                `json:"lessonIndex"`
	Type           string             `json:"type"`
	SubjectID      primitive.ObjectID `json:"subjectID"`
	GroupID        primitive.ObjectID `json:"groupID"`
	TeacherID      primitive.ObjectID `json:"teacherID"`
	RoomID         primitive.ObjectID `json:"roomID"`
}

type UpdateLessonDTO struct {
	AddLessonDTO
	Id          primitive.ObjectID `json:"id" binding:"req"`
	Title       string             `json:"title"`
	Homework    string             `json:"homework"`
	Description string             `json:"description"`
}
