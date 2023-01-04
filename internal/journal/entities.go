package journal

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/general"
	parser "studyum/internal/parser/entities"
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
	ID                 string       `json:"id" bson:"_id"`
	Title              string       `json:"title" bson:"title"`
	Lessons            []*Lesson    `json:"lessons" bson:"lessons"`
	NumericMarksSum    int          `json:"numericMarksSum" bson:"numericMarksSum"`
	NumericMarksLength int          `json:"numericMarksAmount" bson:"numericMarksAmount"`
	AbsencesAmount     int          `json:"absencesAmount" bson:"absencesAmount"`
	AbsencesTime       int          `json:"absencesTime" bson:"absencesTime"`
	MarksAmount        []MarkAmount `json:"marksAmount" bson:"marksAmount"`
	Color              string       `json:"color" bson:"color"`
}

type AvailableOption struct {
	Teacher  string `json:"teacher"`
	Subject  string `json:"subject"`
	Group    string `json:"group"`
	Editable bool   `json:"editable"`
}

type Mark struct {
	ID           primitive.ObjectID    `json:"id" bson:"_id"`
	Mark         string                `json:"mark" bson:"mark"`
	StudentID    primitive.ObjectID    `json:"studentID" bson:"studentID"`
	LessonID     primitive.ObjectID    `json:"lessonID" bson:"lessonID"`
	StudyPlaceID primitive.ObjectID    `json:"studyPlaceID" bson:"studyPlaceID"`
	ParsedInfo   parser.ParsedInfoType `json:"-" bson:"parsedInfo"`
}

type Absence struct {
	ID           primitive.ObjectID    `json:"id" bson:"_id"`
	Time         *int                  `json:"time" bson:"time"`
	StudentID    primitive.ObjectID    `json:"studentID" bson:"studentID"`
	LessonID     primitive.ObjectID    `json:"lessonID" bson:"lessonID"`
	StudyPlaceID primitive.ObjectID    `json:"studyPlaceID" bson:"studyPlaceID"`
	ParsedInfo   parser.ParsedInfoType `json:"-" bson:"parsedInfo"`
}

type MarkAmount struct {
	Mark   string `json:"mark" bson:"mark"`
	Amount int    `json:"amount" bson:"amount"`
}

type Lesson struct {
	Id               primitive.ObjectID    `json:"id" bson:"_id"`
	StudyPlaceId     primitive.ObjectID    `json:"studyPlaceId" bson:"studyPlaceId"`
	PrimaryColor     string                `json:"primaryColor" bson:"primaryColor"`
	JournalCellColor string                `json:"journalCellColor" bson:"journalCellColor"`
	SecondaryColor   string                `json:"secondaryColor" bson:"secondaryColor"`
	Type             string                `json:"type" bson:"type"`
	EndDate          time.Time             `json:"endDate" bson:"endDate"`
	StartDate        time.Time             `json:"startDate" bson:"startDate"`
	Marks            []Mark                `json:"marks" bson:"marks"`
	Absences         []Absence             `json:"absences" bson:"absences"`
	Subject          string                `json:"subject" bson:"subject"`
	Group            string                `json:"group" bson:"group"`
	Teacher          string                `json:"teacher" bson:"teacher"`
	Room             string                `json:"room" bson:"room"`
	Title            string                `json:"title" bson:"title"`
	Homework         string                `json:"homework" bson:"homework"`
	Description      string                `json:"description" bson:"description"`
	IsGeneral        bool                  `json:"isGeneral" bson:"isGeneral"`
	ParsedInfo       parser.ParsedInfoType `json:"-" bson:"parsedInfo"`
}

type GeneratedTable struct {
	Titles []string   `json:"titles" bson:"titles"`
	Rows   [][]string `json:"rows" bson:"rows"`
}
