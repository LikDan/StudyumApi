package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AddGeneralLessonDTO struct {
	PrimaryColor     string             `json:"primaryColor" binding:"hexcolor|eq=transparent"`
	SecondaryColor   string             `json:"secondaryColor" binding:"hexcolor|eq=transparent"`
	LessonIndex      int                `json:"lessonIndex"`
	DayIndex         int                `json:"dayIndex"`
	WeekIndex        int                `json:"weekIndex"`
	StartTimeMinutes int                `json:"startTimeMinutes" binding:"req"`
	EndTimeMinutes   int                `json:"endTimeMinutes" binding:"req"`
	SubjectID        primitive.ObjectID `json:"subjectID"`
	TeacherID        primitive.ObjectID `json:"teacher"`
	GroupID          primitive.ObjectID `json:"group"`
	RoomID           primitive.ObjectID `json:"room"`
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

type AddScheduleInfoDTO struct {
	Status string    `json:"status"`
	Date   time.Time `json:"date"`
}

type UpdateLessonDTO struct {
	AddLessonDTO
	Id          primitive.ObjectID `json:"id" binding:"req"`
	Title       string             `json:"title"`
	Homework    string             `json:"homework"`
	Description string             `json:"description"`
}
