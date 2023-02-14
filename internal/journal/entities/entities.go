package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	general "studyum/internal/general/entities"
	"time"
)

type Journal struct {
	Info  Info     `json:"info" bson:"info"`
	Rows  []Row    `json:"rows" bson:"rows"`
	Dates []Lesson `json:"dates" bson:"dates"`
}

type Info struct {
	Editable   bool               `json:"editable" bson:"editable"`
	StudyPlace general.StudyPlace `json:"studyPlace" bson:"studyPlace"`
	Group      string             `json:"group" bson:"group"`
	Teacher    string             `json:"teacher" bson:"teacher"`
	Subject    string             `json:"subject" bson:"subject"`
}

type Row struct {
	ID                 string         `json:"id" bson:"_id"`
	Title              string         `json:"title" bson:"title"`
	Cells              []*Cell        `json:"cells" bson:"cells"`
	AverageMark        float32        `json:"averageMark" bson:"averageMark"`
	NumericMarksSum    int            `json:"numericMarksSum" bson:"numericMarksSum"`
	NumericMarksLength int            `json:"numericMarksAmount" bson:"numericMarksAmount"`
	AbsencesAmount     int            `json:"absencesAmount" bson:"absencesAmount"`
	AbsencesTime       int            `json:"absencesTime" bson:"absencesTime"`
	MarksAmount        map[string]int `json:"marksAmount" bson:"marksAmount"`
	Color              string         `json:"color" bson:"color"`
}

type Cell struct {
	Id               primitive.ObjectID `json:"id" bson:"_id"`
	Type             []string           `json:"type" bson:"type"`
	JournalCellColor string             `json:"journalCellColor" bson:"journalCellColor"`
	Marks            []Mark             `json:"marks,omitempty" bson:"marks"`
	Absences         []Absence          `json:"absences,omitempty" bson:"absences"`
}

type AvailableOption struct {
	Teacher  string `json:"teacher"`
	Subject  string `json:"subject"`
	Group    string `json:"group"`
	Editable bool   `json:"editable"`
}

type DeleteMarkID struct {
	ID primitive.ObjectID `apps:"trackable,collection=Lessons,type=array,nested=marks"`
}

type Mark struct {
	ID           primitive.ObjectID `json:"id" bson:"_id" apps:"trackable,collection=Lessons,type=array,nested=marks"`
	Mark         string             `json:"mark" bson:"mark"`
	StudentID    primitive.ObjectID `json:"studentID" bson:"studentID"`
	LessonID     primitive.ObjectID `json:"lessonID" bson:"lessonID"`
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
}

type DeleteAbsenceID struct {
	ID primitive.ObjectID `apps:"trackable,collection=Lessons,type=array,nested=absences"`
}

type Absence struct {
	ID           primitive.ObjectID `json:"id" bson:"_id" apps:"trackable,collection=Lessons,type=array,nested=absences"`
	Time         *int               `json:"time" bson:"time"`
	StudentID    primitive.ObjectID `json:"studentID" bson:"studentID"`
	LessonID     primitive.ObjectID `json:"lessonID" bson:"lessonID"`
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
}

type MarkAmount struct {
	Mark   string `json:"mark" bson:"mark"`
	Amount int    `json:"amount" bson:"amount"`
}

type DeleteLessonID struct {
	ID primitive.ObjectID `apps:"trackable,collection=Lessons"`
}

type Lesson struct {
	Id               primitive.ObjectID `json:"id" bson:"_id" apps:"trackable,collection=Lessons"`
	StudyPlaceId     primitive.ObjectID `json:"studyPlaceId" bson:"studyPlaceId"`
	PrimaryColor     string             `json:"primaryColor" bson:"primaryColor"`
	JournalCellColor string             `json:"journalCellColor" bson:"journalCellColor"`
	SecondaryColor   string             `json:"secondaryColor" bson:"secondaryColor"`
	Type             string             `json:"type" bson:"type"`
	EndDate          time.Time          `json:"endDate" bson:"endDate"`
	StartDate        time.Time          `json:"startDate" bson:"startDate"`
	LessonIndex      int                `json:"lessonIndex" bson:"lessonIndex"`
	Marks            []Mark             `json:"marks,omitempty" bson:"marks"`
	Absences         []Absence          `json:"absences,omitempty" bson:"absences"`
	Subject          string             `json:"subject" bson:"subject"`
	Group            string             `json:"group" bson:"group"`
	Teacher          string             `json:"teacher" bson:"teacher"`
	Room             string             `json:"room" bson:"room"`
	Title            string             `json:"title" bson:"title"`
	Homework         string             `json:"homework" bson:"homework"`
	Description      string             `json:"description" bson:"description"`
	IsGeneral        bool               `json:"isGeneral" bson:"isGeneral"`
}

type GeneratedTable struct {
	Titles []string   `json:"titles" bson:"titles"`
	Rows   [][]string `json:"rows" bson:"rows"`
}
