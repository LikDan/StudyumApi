package schedule

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type SubjectFull struct {
	Id               primitive.ObjectID `json:"id" bson:"_id"`
	Subject          string             `json:"subject"`
	Teacher          string             `json:"teacher"`
	Group            string             `json:"group"`
	Room             string             `json:"room"`
	ColumnIndex      int                `json:"columnIndex" bson:"columnIndex"`
	RowIndex         int                `json:"rowIndex" bson:"rowIndex"`
	WeekIndex        int                `json:"weekIndex" bson:"weekIndex"`
	Type_            string             `json:"type" bson:"type"`
	EducationPlaceId int                `json:"educationPlaceId" bson:"educationPlaceId"`
	Date             time.Time          `json:"date"`
	Homework         string             `json:"homework"`
	SmallDescription string             `json:"smallDescription"`
	Description      string             `json:"description"`
}

func subjectToBsonWithoutType(subject SubjectFull) bson.D {
	return bson.D{
		{"columnIndex", subject.ColumnIndex},
		{"rowIndex", subject.RowIndex},
		{"weekIndex", subject.WeekIndex},
		{"subject", subject.Subject},
		{"teacher", subject.Teacher},
		{"group", subject.Group},
		{"room", subject.Room},
		{"educationPlaceId", subject.EducationPlaceId},
	}
}
