package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
)

func getSchedule(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)

	type_ := ctx.Query("type")
	name := ctx.Query("name")
	studyPlaceIdStr := ctx.Query("studyPlaceId")

	if err == nil {
		if type_ == "" {
			type_ = user["type"].(string)
		}
		if name == "" {
			name = user["name"].(string)
		}
		if studyPlaceIdStr == "" {
			studyPlaceIdStr = strconv.Itoa(int(user["studyPlaceId"].(int32)))
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
	if err != nil {
		println(err.Error())
	}

	stateCursor, err := stateCollection.Find(
		nil,
		bson.D{{"educationPlaceId", educationPlaceId}},
		options.Find().SetSort(bson.D{{"weekIndex", 1}, {"dayIndex", 1}}),
	)
	checkError(err)

	var states []StateInfo

	for stateCursor.TryNext(nil) {
		weekIndex := stateCursor.Current.Lookup("weekIndex").Int32()
		dayIndex := stateCursor.Current.Lookup("dayIndex").Int32()
		state := State(stateCursor.Current.Lookup("status").StringValue())

		stateInfo := StateInfo{
			State:        state,
			WeekIndex:    int(weekIndex),
			DayIndex:     int(dayIndex),
			StudyPlaceId: educationPlaceId,
		}

		states = append(states, stateInfo)
	}

	lessonsCursor, err := subjectsCollection.Aggregate(nil, mongo.Pipeline{
		bson.D{{"$match", bson.M{type_: name, "educationPlaceId": educationPlaceId}}},
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

	if err != nil {
		println(err.Error())
	}

	var lessons []*Lesson

	err = lessonsCursor.All(nil, &lessons)
	if err != nil {
		println(err.Error())
	}

	for i := 0; i < studyPlace.SubjectsQuantity*studyPlace.DaysQuantity*studyPlace.WeeksQuantity; i++ {
		if len(lessons) <= i {
			lessons = append(lessons, nil)
			continue
		}

		if lessons[i].Id == i {
			lessons[i].IsStay = true

			for _, subject := range lessons[i].Subjects {
				if subject.Type_ != "STAY" {
					lessons[i].IsStay = false
					break
				}
			}

			continue
		}

		lessons = append(lessons[:i+1], lessons[i:]...)
		lessons[i] = nil
	}

	ctx.JSON(200, gin.H{
		"status":   states,
		"subjects": lessons,
		"info": gin.H{
			"weeksCount":     studyPlace.WeeksQuantity,
			"daysCount":      studyPlace.DaysQuantity,
			"subjectsCount":  studyPlace.SubjectsQuantity,
			"type_":          type_,
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
