package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	core "studyum/internal/core/entities"
	"time"
)

type Journal struct {
	Cells     []JournalCell     `json:"cells" bson:"cells"`
	Dates     []JournalDate     `json:"dates" bson:"dates"`
	RowTitles []JournalRowTitle `json:"rowTitles" bson:"rowTitles"`
	Info      JournalInfo       `json:"info" bson:"info"`
}

type JournalDate struct {
	ID      primitive.ObjectID   `json:"id" bson:"_id"`
	Date    time.Time            `json:"date" bson:"date"`
	TypeIDs []primitive.ObjectID `json:"typeIDs" bson:"typeIDs"`
}

type JournalRowTitle struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Title string             `json:"title" bson:"title"`
}

type JournalCell struct {
	Date     JournalDate        `json:"-" bson:"date"`
	RowTitle JournalRowTitle    `json:"-" bson:"rowTitle"`
	Entries  []JournalCellEntry `json:"entries" bson:"entries"`
	Point    Point              `json:"point" bson:"point"`
}

type JournalCellEntry struct {
	LessonsID primitive.ObjectID `json:"lessonID" bson:"lessonID"`
	TypeID    primitive.ObjectID `json:"typeID" bson:"typeID"`
	Marks     []Mark             `json:"marks" bson:"marks"`
	Absences  []Absence          `json:"absences" bson:"absences"`
}

type Point struct {
	X int `json:"x" bson:"x"`
	Y int `json:"y" bson:"y"`
}

type AvailableOption struct {
	Teacher    core.Teacher `json:"teacher"`
	Subject    core.Subject `json:"subject"`
	Group      core.Group   `json:"group"`
	Header     string       `json:"header"`
	Editable   bool         `json:"editable"`
	HasMembers bool         `json:"hasMembers"`
}

type CategoryOptions struct {
	Category string            `json:"category"`
	Options  []AvailableOption `json:"options"`
}

type DeleteMarkID struct {
	ID primitive.ObjectID `apps:"trackable,collection=Lessons,type=array,nested=marks"`
}

type StudentMark struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	MarkID       primitive.ObjectID `json:"markID" bson:"markID"`
	LessonID     primitive.ObjectID `json:"lessonID" bson:"lessonID"`
	StudentID    primitive.ObjectID `json:"studentID" bson:"studentID"`
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
}

type Mark struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	MarkID       primitive.ObjectID `json:"markID" bson:"markID"`
	Mark         string             `json:"mark" bson:"mark"`
	MarkWeight   int                `json:"markWeight" bson:"markWeight"`
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

type JournalLesson struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	StudyPlaceID   string             `json:"studyPlaceID" bson:"studyPlaceID"`
	EndDate        time.Time          `json:"endDate" bson:"endDate"`
	StartDate      time.Time          `json:"startDate" bson:"startDate"`
	PrimaryColor   string             `json:"primaryColor" bson:"primaryColor"`
	SecondaryColor string             `json:"secondaryColor" bson:"secondaryColor"`
	LessonIndex    int                `json:"lessonIndex" bson:"lessonIndex"`
	Subject        core.Subject       `json:"subject" bson:"subject"`
	Group          core.Group         `json:"group" bson:"group"`
	Teacher        core.Teacher       `json:"teacher" bson:"teacher"`
	Room           core.Room          `json:"room" bson:"room"`
	Type           LessonType         `json:"type" bson:"type"`
	SubjectID      primitive.ObjectID `json:"subjectID" bson:"subjectID"`
	GroupID        primitive.ObjectID `json:"groupID" bson:"groupID"`
	TeacherID      primitive.ObjectID `json:"teacherID" bson:"teacherID"`
	RoomID         primitive.ObjectID `json:"roomID" bson:"roomID"`
	TypeID         primitive.ObjectID `json:"typeID" bson:"typeID"`
	Title          string             `json:"title" bson:"title"`
	Homework       string             `json:"homework" bson:"homework"`
	Description    string             `json:"description" bson:"description"`
	Marks          []Mark             `json:"marks" bson:"marks"`
	Absence        *Absence           `json:"absence" bson:"absence"`
}

type LessonType struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	AbsenceMark    string             `json:"absenceMark" bson:"absenceMark"`
	AssignedColor  string             `json:"assignedColor" bson:"assignedColor"`
	StudyPlaceID   primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
	AvailableMarks []Mark             `json:"availableMarks" bson:"availableMarks"`
	Type           string             `json:"type" bson:"type"`
}

type AvailableMark struct {
	ID                  primitive.ObjectID   `json:"id" bson:"_id"`
	AssignLessonTypeIDs []primitive.ObjectID `json:"assignLessonTypeIDs" bson:"assignLessonTypeIDs"`
	Mark                string               `json:"mark" bson:"mark"`
	MarkWeight          int                  `json:"markWeight" bson:"markWeight"`
	StudyPlaceID        primitive.ObjectID   `json:"studyPlaceID" bson:"studyPlaceID"`
}

type Lesson struct {
	Id               primitive.ObjectID `json:"id" bson:"_id" apps:"trackable,collection=Lessons"`
	StudyPlaceId     primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
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
	SubjectID        primitive.ObjectID `json:"subjectID" bson:"subjectID"`
	GroupID          primitive.ObjectID `json:"groupID" bson:"groupID"`
	TeacherID        primitive.ObjectID `json:"teacherID" bson:"teacherID"`
	RoomID           primitive.ObjectID `json:"roomID" bson:"roomID"`
	Title            string             `json:"title" bson:"title"`
	Homework         string             `json:"homework" bson:"homework"`
	Description      string             `json:"description" bson:"description"`
	IsGeneral        bool               `json:"isGeneral" bson:"isGeneral"`
}

type GeneratedTable struct {
	Titles []string   `json:"titles" bson:"titles"`
	Rows   [][]string `json:"rows" bson:"rows"`
}

type JournalInfo struct {
	Editable bool            `json:"editable" bson:"editable"`
	Configs  []JournalConfig `json:"configs" bson:"configs"`
}

type JournalConfig struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Title        string             `json:"title" bson:"title"`
	Index        int                `json:"index" bson:"index"`
	MarkIDs      []string           `json:"markIDs" bson:"markIDs"`
	ShowAbsences bool               `json:"showAbsences" bson:"showAbsences"`
	ShowLatency  bool               `json:"showLatency" bson:"showLatency"`
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID" bson:"studyPlaceID"`
	TypeIDs      []string           `json:"typeIDs" bson:"typeIDs"`
}
