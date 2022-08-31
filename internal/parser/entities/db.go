package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type GeneralLesson struct {
	Id           primitive.ObjectID `bson:"_id"`
	StudyPlaceId int                `bson:"studyPlaceId"`
	EndTime      string             `bson:"endTime"`
	StartTime    string             `bson:"startTime"`
	Subject      string             `bson:"subject"`
	Group        string             `bson:"group"`
	Teacher      string             `bson:"teacher"`
	Room         string             `bson:"room"`
	DayIndex     int                `bson:"dayIndex"`
	WeekIndex    int                `bson:"weekIndex"`
	ParsedInfo   ParsedInfoType     `bson:"parsedInfo"`
}

type Lesson struct {
	Id             primitive.ObjectID `bson:"_id"`
	StudyPlaceId   int                `bson:"studyPlaceId"`
	PrimaryColor   string             `bson:"primaryColor"`
	SecondaryColor string             `bson:"secondaryColor"`
	EndDate        time.Time          `bson:"endDate"`
	StartDate      time.Time          `bson:"startDate"`
	Subject        string             `bson:"subject"`
	Group          string             `bson:"group"`
	Teacher        string             `bson:"teacher"`
	Room           string             `bson:"room"`
	Marks          []Mark             `bson:"marks"`
	Title          string             `bson:"title"`
	Homework       string             `bson:"homework"`
	Description    string             `bson:"description"`
	ParsedInfo     ParsedInfoType     `bson:"parsedInfo"`
}

type Mark struct {
	Id           primitive.ObjectID `bson:"_id"`
	Mark         string             `bson:"mark"`
	StudentID    primitive.ObjectID `bson:"studentID"`
	LessonId     primitive.ObjectID `bson:"lessonId"`
	StudyPlaceId int                `bson:"studyPlaceId"`
	ParsedInfo   ParsedInfoType     `bson:"parsedInfo"`
}

type User struct {
	Id            primitive.ObjectID `json:"id" bson:"_id"`
	Token         string             `json:"-" bson:"token"`
	Password      string             `json:"password" bson:"password"`
	Email         string             `json:"email" bson:"email"`
	FirebaseToken string             `json:"-" bson:"firebaseToken"`
	VerifiedEmail bool               `json:"verifiedEmail" bson:"verifiedEmail"`
	Login         string             `json:"login" bson:"login"`
	Name          string             `json:"name" bson:"name"`
	PictureUrl    string             `json:"picture" bson:"picture"`
	Type          string             `json:"type" bson:"type"`
	TypeName      string             `json:"typeName" bson:"typeName"`
	StudyPlaceId  int                `json:"studyPlaceId" bson:"studyPlaceId"`
	Permissions   []string           `json:"permissions" bson:"permissions"`
	Accepted      bool               `json:"accepted" bson:"accepted"`
	Blocked       bool               `json:"blocked" bson:"blocked"`
}
