package schedule

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
	h "studyum/src/api"
	"studyum/src/db"
	"time"
)

func getScheduleOld(ctx *gin.Context) {
	type_ := ctx.Query("type")
	name := ctx.Query("name")
	studyPlaceIdStr := ctx.Query("studyPlaceId")

	if type_ == "" || name == "" || studyPlaceIdStr == "" {
		h.ErrorMessage(ctx, "not authorized")
		return
	}

	educationPlaceId, err := strconv.Atoi(studyPlaceIdStr)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, "not valid params")
		return
	}

	var studyPlace StudyPlaceOld

	err = db.StudyPlacesCollection.FindOne(nil, bson.M{"_id": educationPlaceId}).Decode(&studyPlace)
	if h.CheckError(err, h.WARNING) {
		return
	}

	stateCursor, err := db.StateCollection.Find(
		nil,
		bson.D{{"educationPlaceId", educationPlaceId}},
		options.Find().SetSort(bson.D{{"weekIndex", 1}, {"dayIndex", 1}}),
	)
	h.CheckError(err, h.WARNING)

	var states []StateInfo
	err = stateCursor.All(nil, &states)
	if h.CheckError(err, h.WARNING) {
		return
	}

	startDate := h.Date().AddDate(0, 0, 1-int(time.Now().Weekday()))

	lessonsCursor, err := db.SubjectsCollection.Aggregate(nil, mongo.Pipeline{
		bson.D{{"$match", bson.M{"date": bson.M{"$gte": startDate}, type_: name, "educationPlaceId": educationPlaceId}}},
		bson.D{{"$group", bson.M{
			"_id":         bson.M{"$sum": bson.A{bson.M{"$multiply": bson.A{"$weekIndex", studyPlace.DaysQuantity, studyPlace.SubjectsQuantity}}, bson.M{"$multiply": bson.A{"$columnIndex", studyPlace.SubjectsQuantity}}, "$rowIndex"}},
			"weekIndex":   bson.M{"$first": "$weekIndex"},
			"columnIndex": bson.M{"$first": "$columnIndex"},
			"rowIndex":    bson.M{"$first": "$rowIndex"},
			"date":        bson.M{"$first": "$date"},
			"subjects":    bson.M{"$addToSet": bson.M{"subject": "$subject", "group": "$group", "teacher": "$teacher", "room": "$room", "type": "$type"}},
		}}},
		bson.D{{"$sort", bson.M{"_id": 1}}},
	})

	if h.CheckError(err, h.WARNING) {
		return
	}

	var lessons []*LessonOld

	err = lessonsCursor.All(nil, &lessons)
	if h.CheckError(err, h.WARNING) {
		return
	}

	lastLesson := lessons[len(lessons)-1]

	_, currentWeekIndex := time.Now().ISOWeek()
	currentWeekIndex %= studyPlace.WeeksQuantity

	lessonsCursor, err = db.GeneralSubjectsCollection.Aggregate(nil, mongo.Pipeline{
		bson.D{{"$match", bson.M{type_: name, "studyPlaceId": educationPlaceId, "$or": bson.A{bson.M{"weekIndex": bson.M{"$ne": currentWeekIndex}}, bson.M{"$and": bson.A{bson.M{"weekIndex": bson.M{"$eq": lastLesson.WeekIndex}}, bson.M{"dayIndex": bson.M{"$gt": lastLesson.ColumnIndex}}}}}}}},
		bson.D{{"$group", bson.M{
			"_id":       bson.M{"$sum": bson.A{bson.M{"$multiply": bson.A{"$weekIndex", studyPlace.DaysQuantity, studyPlace.SubjectsQuantity}}, bson.M{"$multiply": bson.A{"$columnIndex", studyPlace.SubjectsQuantity}}, "$rowIndex"}},
			"weekIndex": bson.M{"$first": "$weekIndex"},
			"dayIndex":  bson.M{"$first": "$dayIndex"},
			"rowIndex":  bson.M{"$first": "$rowIndex"},
			"date":      bson.M{"$first": "$date"},
			"subjects":  bson.M{"$addToSet": bson.M{"subject": "$subject", "group": "$group", "teacher": "$teacher", "room": "$room", "type": "$type"}},
		}}},
		bson.D{{"$sort", bson.M{"_id": 1}}},
	})
	if h.CheckError(err, h.WARNING) {
		return
	}

	var generalLessons []*LessonOld
	err = lessonsCursor.All(nil, &generalLessons)
	if h.CheckError(err, h.WARNING) {
		return
	}

	for i := 0; i < studyPlace.SubjectsQuantity*studyPlace.DaysQuantity*studyPlace.WeeksQuantity; i++ {
		if len(generalLessons) <= i {
			generalLessons = append(generalLessons, nil)
			continue
		}

		if generalLessons[i].Id == i {
			generalLessons[i].IsStay = true

			for _, subject := range generalLessons[i].Subjects {
				subject.Type_ = "STAY"
			}

			continue
		}
		generalLessons = append(generalLessons[:i+1], generalLessons[i:]...)
		generalLessons[i] = nil
	}

	for _, lesson := range lessons {
		generalLessons[lesson.Id] = lesson

		lesson.IsStay = true
		for _, subject := range lesson.Subjects {
			if subject.Type_ != "STAY" {
				lesson.IsStay = false
				break
			}
		}
	}

	currentWeekStartIndex := currentWeekIndex * studyPlace.DaysQuantity * studyPlace.SubjectsQuantity
	generalLessons = append(generalLessons[currentWeekStartIndex:], generalLessons[:currentWeekStartIndex]...)

	ctx.JSON(200, gin.H{
		"status":   states,
		"subjects": generalLessons,
		"info": gin.H{
			"weeksCount":     studyPlace.WeeksQuantity,
			"daysCount":      studyPlace.DaysQuantity,
			"subjectsCount":  studyPlace.SubjectsQuantity,
			"type":           type_,
			"name":           name,
			"studyPlaceId":   educationPlaceId,
			"studyPlaceName": studyPlace.Name,
		},
	})
}

type StudyPlaceOld struct {
	Id               int32  `json:"id" bson:"_id"`
	WeeksQuantity    int    `json:"weeksQuantity" bson:"weeksCount"`
	DaysQuantity     int    `json:"daysQuantity" bson:"daysCount"`
	SubjectsQuantity int    `json:"subjectsQuantity" bson:"subjectsCount"`
	Name             string `json:"name" bson:"name"`
}

type LessonOld struct {
	Id          int           `bson:"_id" json:"-"`
	Subjects    []*SubjectOld `bson:"subjects" json:"subjects"`
	ColumnIndex int32         `bson:"columnIndex" json:"columnIndex"`
	RowIndex    int32         `bson:"rowIndex" json:"rowIndex"`
	WeekIndex   int32         `bson:"weekIndex" json:"weekIndex"`
	Date        time.Time     `bson:"date" json:"-"`
	IsStay      bool          `bson:"isStay" json:"isStay"`
}

type SubjectOld struct {
	Subject string `bson:"subject" json:"subject"`
	Teacher string `bson:"teacher" json:"teacher"`
	Group   string `bson:"group" json:"group"`
	Room    string `bson:"room" json:"room"`
	Type_   string `bson:"type" json:"type"`
}

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
	StartTime        time.Time          `json:"startTime" bson:"startTime"`
	EndTime          time.Time          `json:"endTime" bson:"endTime"`
}

type State string

const (
	Updated    State = "UPDATED"
	NotUpdated State = "NOT_UPDATED"
)

type StateInfo struct {
	State        State `bson:"status" json:"status"`
	WeekIndex    int   `bson:"weekIndex" json:"weekIndex"`
	DayIndex     int   `bson:"dayIndex" json:"dayIndex"`
	StudyPlaceId int   `bson:"educationPlaceId" json:"-"`
}

func getScheduleTypesOld(ctx *gin.Context) {
	var res []string

	educationPlaceIdStr := ctx.Query("studyPlaceId")
	if educationPlaceIdStr == "" {
		h.ErrorMessage(ctx, "provide all params")
		return
	}

	educationPlaceId, err := strconv.Atoi(educationPlaceIdStr)
	h.CheckError(err, h.UNDEFINED)

	var toJson = func(type_ string) {
		var filter = bson.D{{type_, bson.D{{"$not", bson.D{{"$eq", ""}}}}}, {"educationPlaceId", bson.D{{"$eq", educationPlaceId}}}}
		types, _ := db.SubjectsCollection.Distinct(nil, type_, filter)

		for _, response := range types {
			res = append(res, "{\"type\": \""+type_+"\",\"name\": \""+response.(string)+"\"}")
		}
	}

	toJson("room")
	toJson("group")
	toJson("teacher")
	toJson("subject")

	_, err = fmt.Fprintf(ctx.Writer, "[%s]", strings.Join(res, ", "))
	h.CheckError(err, h.WARNING)
}

func BuildRequests(api *gin.RouterGroup) {
	api.GET("", getScheduleOld)
	api.GET("/types", getScheduleTypesOld)
}
