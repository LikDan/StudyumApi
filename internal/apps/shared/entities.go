package shared

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Data map[string]any

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name" encryption:""`
	Type         string             `json:"type" bson:"type"`
	TypeName     string             `json:"typeName" bson:"typename"`
	TuitionGroup string             `json:"tuitionGroup" bson:"tuitionGroup"`
	Data         Data               `json:"data" bson:"appData"`
}

type Lesson struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	PrimaryColor     string             `json:"primaryColor" bson:"primaryColor"`
	JournalCellColor string             `json:"journalCellColor" bson:"journalCellColor"`
	SecondaryColor   string             `json:"secondaryColor" bson:"secondaryColor"`
	Type             string             `json:"type" bson:"type"`
	EndDate          time.Time          `json:"endDate" bson:"endDate"`
	StartDate        time.Time          `json:"startDate" bson:"startDate"`
	LessonIndex      int                `json:"lessonIndex" bson:"lessonIndex"`
	Subject          string             `json:"subject" bson:"subject"`
	Group            string             `json:"group" bson:"group"`
	Teacher          string             `json:"teacher" bson:"teacher"`
	Room             string             `json:"room" bson:"room"`
	Title            string             `json:"title" bson:"title"`
	Homework         string             `json:"homework" bson:"homework"`
	Description      string             `json:"description" bson:"description"`
	Data             Data               `json:"data" bson:"appData"`
}

type Mark struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Mark      string             `json:"mark" bson:"mark"`
	StudentID primitive.ObjectID `json:"studentID" bson:"studentID"`
	LessonID  primitive.ObjectID `json:"lessonID" bson:"lessonID"`
}
