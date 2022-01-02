package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func getSchedule(w http.ResponseWriter, r *http.Request) {
	log.Println("GET SCHEDULE")
	type_, err := getUrlData(r, "type")
	if checkError(err) {
		return
	}

	name, err := getUrlData(r, "name")
	if checkError(err) {
		return
	}

	educationPlaceIdStr, err := getUrlData(r, "educationPlaceId")
	if checkError(err) {
		return
	}

	educationPlaceId, err := strconv.Atoi(educationPlaceIdStr)
	if checkError(err) {
		return
	}

	var educationPlace bson.M

	educationPlaceResult := studyPlacesCollection.FindOne(nil, bson.M{"_id": educationPlaceId})
	err = educationPlaceResult.Decode(&educationPlace)
	checkError(err)

	educationPlaceName := educationPlace["name"].(string)
	weeksAmount := educationPlace["weeksCount"].(int32)
	daysAmount := educationPlace["daysCount"].(int32)
	subjectsAmount := educationPlace["subjectsCount"].(int32)

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
			state:            state,
			weekIndex:        int(weekIndex),
			dayIndex:         int(dayIndex),
			educationPlaceId: educationPlaceId,
		}

		states = append(states, stateInfo)
	}

	lessonsCursor, err := subjectsCollection.Aggregate(nil, mongo.Pipeline{
		bson.D{{"$match", bson.D{{type_, name}, {"educationPlaceId", educationPlaceId}}}},
		bson.D{{"$group", bson.D{
			{"_id", bson.D{{"weekIndex", "$weekIndex"}, {"columnIndex", "$columnIndex"}, {"rowIndex", "$rowIndex"}}},
			{"subjects", bson.D{{"$addToSet", bson.D{{"subject", "$subject"}, {"group", "$group"}, {"teacher", "$teacher"}, {"room", "$room"}, {"type", "$type"}}}}},
		}}},
		bson.D{{"$sort", bson.D{{"_id.weekIndex", 1}, {"_id.columnIndex", 1}, {"_id.rowIndex", 1}}}},
	})
	checkError(err)

	currentRowIndex := int32(-1)
	currentColumnIndex := int32(0)
	currentWeekIndex := int32(0)

	add := func() {
		currentRowIndex++
		if currentRowIndex >= subjectsAmount {
			currentRowIndex = 0
			currentColumnIndex++
		}
		if currentColumnIndex >= daysAmount {
			currentColumnIndex = 0
			currentWeekIndex++
		}
	}

	var lessons []*Lesson

	for lessonsCursor.TryNext(nil) {
		var subjects []Subject

		lessonRaw := lessonsCursor.Current

		add()

		weekIndex := lessonRaw.Lookup("_id", "weekIndex").Int32()
		columnIndex := lessonRaw.Lookup("_id", "columnIndex").Int32()
		rowIndex := lessonRaw.Lookup("_id", "rowIndex").Int32()

		for currentRowIndex != rowIndex || currentColumnIndex != columnIndex || currentWeekIndex != weekIndex {
			add()
			lessons = append(lessons, nil)
		}

		subjectsRaw, _ := lessonRaw.Lookup("subjects").Array().Values()
		for _, subjectRaw := range subjectsRaw {
			subjectDoc := subjectRaw.Document()

			subjectName := subjectDoc.Lookup("subject").StringValue()
			teacher := subjectDoc.Lookup("teacher").StringValue()
			group := subjectDoc.Lookup("group").StringValue()
			room := subjectDoc.Lookup("room").StringValue()
			type_ := subjectDoc.Lookup("type").StringValue()

			subject := Subject{
				subject: subjectName,
				teacher: teacher,
				group:   group,
				room:    room,
				type_:   type_,
			}

			subjects = append(subjects, subject)
		}

		lesson := &Lesson{
			subjects:    subjects,
			columnIndex: columnIndex,
			rowIndex:    rowIndex,
			weekIndex:   weekIndex,
		}

		lessons = append(lessons, lesson)
	}

	for currentRowIndex != subjectsAmount-1 || currentColumnIndex != daysAmount-1 || currentWeekIndex != weeksAmount-1 {
		add()
		lessons = append(lessons, nil)
	}

	var lessonsJson []string
	var statesJson []string

	for _, lesson := range lessons {
		if lesson == nil {
			lessonsJson = append(lessonsJson, "null")
			continue
		}
		lessonsJson = append(lessonsJson, lesson.toJson())
	}

	for _, state := range states {
		statesJson = append(statesJson, state.toJsonWithoutId())
	}

	_, err = fmt.Fprintln(w, "{\"status\": ["+strings.Join(statesJson, ", ")+
		"], \"subjects\": ["+strings.Join(lessonsJson, ", ")+
		"], \"info\": {"+
		"\"weeksCount\": "+strconv.Itoa(int(weeksAmount))+
		", \"daysCount\": "+strconv.Itoa(int(daysAmount))+
		", \"subjectsCount\": "+strconv.Itoa(int(subjectsAmount))+
		", \"type\": \""+type_+
		"\", \"name\": \""+name+
		"\", \"educationPlaceId\": "+educationPlaceIdStr+
		", \"educationPlaceName\": \""+educationPlaceName+"\"}}")
	checkError(err)
}

func getScheduleTypes(w http.ResponseWriter, r *http.Request) {
	var res []string

	educationPlaceIdStr, err := getUrlData(r, "educationPlaceId")
	if err != nil {
		_, err = fmt.Fprintf(w, "Please provide all params (url: %s)", r.URL.Path)
		checkError(err)

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

	_, err = fmt.Fprintf(w, "[%s]", strings.Join(res, ", "))
	checkError(err)
}
