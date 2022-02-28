package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
	"time"
)

func getSchedule(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)

	type_ := ctx.Query("type")
	name := ctx.Query("name")
	studyPlaceIdStr := ctx.Query("studyPlaceId")

	if err == nil {
		if type_ == "" {
			type_ = user.Type
		}
		if name == "" {
			name = user.Name
		}
		if studyPlaceIdStr == "" {
			studyPlaceIdStr = strconv.Itoa(user.StudyPlaceId)
		}
	}

	if type_ == "" || name == "" || studyPlaceIdStr == "" {
		errorMessage(ctx, "not authorized")
		return
	}

	educationPlaceId, err := strconv.Atoi(studyPlaceIdStr)
	if checkError(err) {
		errorMessage(ctx, "not valid params")
		return
	}

	var studyPlace StudyPlace

	err = studyPlacesCollection.FindOne(nil, bson.M{"_id": educationPlaceId}).Decode(&studyPlace)
	if checkError(err) {
		return
	}

	stateCursor, err := stateCollection.Find(
		nil,
		bson.D{{"educationPlaceId", educationPlaceId}},
		options.Find().SetSort(bson.D{{"weekIndex", 1}, {"dayIndex", 1}}),
	)
	checkError(err)

	var states []StateInfo
	err = stateCursor.All(nil, &states)
	if checkError(err) {
		return
	}

	startDate := Date().AddDate(0, 0, 1-int(time.Now().Weekday()))

	lessonsCursor, err := subjectsCollection.Aggregate(nil, mongo.Pipeline{
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

	if checkError(err) {
		return
	}

	var lessons []*Lesson

	err = lessonsCursor.All(nil, &lessons)
	if checkError(err) {
		return
	}

	lastLesson := lessons[len(lessons)-1]

	_, currentWeekIndex := time.Now().ISOWeek()
	currentWeekIndex %= studyPlace.WeeksQuantity

	lessonsCursor, err = generalSubjectsCollection.Aggregate(nil, mongo.Pipeline{
		bson.D{{"$match", bson.M{type_: name, "educationPlaceId": educationPlaceId, "$or": bson.A{bson.M{"weekIndex": bson.M{"$ne": currentWeekIndex}}, bson.M{"$and": bson.A{bson.M{"weekIndex": bson.M{"$eq": lastLesson.WeekIndex}}, bson.M{"columnIndex": bson.M{"$gt": lastLesson.ColumnIndex}}}}}}}},
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
	if checkError(err) {
		return
	}

	var generalLessons []*Lesson
	err = lessonsCursor.All(nil, &generalLessons)
	if checkError(err) {
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

func getScheduleTypes(ctx *gin.Context) {
	var res []string

	educationPlaceIdStr := ctx.Query("studyPlaceId")
	if educationPlaceIdStr == "" {
		errorMessage(ctx, "provide all params")
		return
	}

	educationPlaceId, err := strconv.Atoi(educationPlaceIdStr)
	checkError(err)

	var toJson = func(type_ string) {
		var filter = bson.D{{type_, bson.D{{"$not", bson.D{{"$eq", ""}}}}}, {"educationPlaceId", bson.D{{"$eq", educationPlaceId}}}}
		types, _ := subjectsCollection.Distinct(nil, type_, filter)

		for _, response := range types {
			res = append(res, "{\"type\": \""+type_+"\", \"name\": \""+response.(string)+"\"}")
		}
	}

	toJson("room")
	toJson("group")
	toJson("teacher")
	toJson("subject")

	_, err = fmt.Fprintf(ctx.Writer, "[%s]", strings.Join(res, ", "))
	checkError(err)
}

func updateSchedule(ctx *gin.Context) {
	edu, err := getEducationViaPasswordRequest(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	UpdateDbSchedule(edu)
}
