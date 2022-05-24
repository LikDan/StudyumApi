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
	h "studyium/src/api"
	userApi "studyium/src/api/user"
	"studyium/src/db"
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

func getSchedule(ctx *gin.Context) {
	var user userApi.User
	if err := userApi.GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	type_ := ctx.DefaultQuery("type", user.Type)
	typeName := ctx.DefaultQuery("name", user.TypeName)

	if !h.CheckNotEmpty(type_, typeName) {
		h.ErrorMessage(ctx, "Provide valid params")
		return
	}

	var schedule Schedule

	startWeekDate := h.Date().AddDate(0, 0, 1-int(time.Now().Weekday()))
	cursor, err := db.StudyPlacesCollection.Aggregate(nil, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": user.StudyPlaceId,
			},
		}, bson.M{
			"$addFields": bson.M{
				"date": bson.M{"$range": bson.A{0, bson.M{"$multiply": bson.A{7, "$weeksCount"}}}},
			},
		}, bson.M{
			"$unwind": "$date",
		}, bson.M{
			"$addFields": bson.M{
				"date": bson.M{"$dateAdd": bson.M{
					"startDate": startWeekDate,
					"unit":      "day",
					"amount":    "$date",
				}},
			},
		}, bson.M{
			"$addFields": bson.M{
				"indexes": bson.M{
					"weekIndex": bson.M{"$mod": bson.A{bson.M{"$isoWeek": "$date"}, "$weeksCount"}},
					"dayIndex":  bson.M{"$subtract": bson.A{bson.M{"$isoDayOfWeek": "$date"}, 1}},
				},
			},
		}, bson.M{
			"$lookup": bson.M{
				"from": "General",
				"let":  bson.M{"weekIndex": "$indexes.weekIndex", "dayIndex": "$indexes.dayIndex", "date": "$date"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$" + type_, typeName}},
									bson.M{"$eq": bson.A{"$weekIndex", "$$weekIndex"}},
									bson.M{"$eq": bson.A{"$dayIndex", "$$dayIndex"}},
								},
							},
						},
					}, bson.M{
						"$addFields": bson.M{
							"updated":   false,
							"type":      "STAY",
							"startDate": bson.M{"$toDate": bson.M{"$concat": bson.A{bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$$date"}}, "T", "$startTime"}}},
							"endDate":   bson.M{"$toDate": bson.M{"$concat": bson.A{bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$$date"}}, "T", "$endTime"}}},
						},
					},
				},
				"as": "general",
			},
		}, bson.M{
			"$lookup": bson.M{
				"from": "Subjects",
				"let":  bson.M{"date": "$date"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$startTime"}}, bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$$date"}}}},
									bson.M{"$eq": bson.A{"$" + type_, typeName}},
								},
							},
						},
					}, bson.M{
						"$addFields": bson.M{
							"updated":   true,
							"startDate": "$startTime",
							"endDate":   "$endTime",
						},
					},
				},
				"as": "lessons",
			},
		}, bson.M{
			"$addFields": bson.M{
				"lessons": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$lessons", bson.A{}}}, "$general", "$lessons"}},
			},
		}, bson.M{
			"$unwind": "$lessons",
		}, bson.M{
			"$group": bson.M{
				"_id": nil,
				"studyPlace": bson.M{"$first": bson.M{
					"_id":        "$_id",
					"name":       "$name",
					"weeksCount": "$weeksCount",
				}},
				"lessons": bson.M{"$push": "$lessons"},
			},
		}, bson.M{
			"$sort": bson.M{"lessons.startDate": 1},
		}, bson.M{
			"$addFields": bson.M{
				"info": bson.M{
					"startWeekDate": startWeekDate,
					"date":          time.Now(),
					"type":          type_,
					"typeName":      typeName,
					"studyPlace":    "$studyPlace",
				},
			},
		},
	})
	if h.CheckAndMessage(ctx, 418, err, h.WARNING) {
		return
	}

	cursor.Next(nil)
	if err = cursor.Decode(&schedule); h.CheckAndMessage(ctx, 418, err, h.WARNING) {
		return
	}

	ctx.JSON(200, schedule)
}

func addLessons(ctx *gin.Context) {
	var user userApi.User
	if err := userApi.GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	if !h.SliceContains(user.Permissions, "editSchedule") {
		h.ErrorMessage(ctx, "no permissions")
		return
	}

	var subjects []*SubjectFull
	if err := ctx.BindJSON(&subjects); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}
	for _, subject := range subjects {
		subject.Id = primitive.NewObjectID()
	}

	if _, err := db.SubjectsCollection.InsertMany(nil, h.ToInterfaceSlice(subjects)); h.CheckAndMessage(ctx, 418, err, h.WARNING) {
		return
	}
	getSchedule(ctx)
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

func getTypes(ctx *gin.Context) {
	var user userApi.User
	if err := userApi.GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	var get = func(type_ string) []string {
		namesInterface, _ := db.SubjectsCollection.Distinct(nil, type_, bson.M{"educationPlaceId": user.StudyPlaceId})

		names := make([]string, len(namesInterface))
		for i, v := range namesInterface {
			names[i] = v.(string)
		}

		return names
	}

	types := Types{
		Groups:   get("group"),
		Teachers: get("teacher"),
		Subjects: get("subject"),
		Rooms:    get("room"),
	}

	ctx.JSON(200, types)
}

type Info struct {
	Type          string     `json:"type" bson:"type"`
	TypeName      string     `json:"typeName" bson:"typeName"`
	StudyPlace    StudyPlace `json:"studyPlace" bson:"studyPlace"`
	StartWeekDate time.Time  `json:"startWeekDate" bson:"startWeekDate"`
	Date          time.Time  `json:"date" bson:"date"`
}

type StudyPlace struct {
	Id         int    `json:"id" bson:"_id"`
	WeeksCount int    `json:"weeksCount" bson:"weeksCount"`
	DaysCount  int    `json:"daysCount" bson:"daysCount"`
	Name       string `json:"name" bson:"name"`
}

type Schedule struct {
	Info    Info      `json:"info" bson:"info"`
	Lessons []*Lesson `json:"lessons" bson:"lessons"`
}

type Types struct {
	Groups   []string `json:"groups" bson:"groups"`
	Teachers []string `json:"teachers" bson:"teachers"`
	Subjects []string `json:"subjects" bson:"subjects"`
	Rooms    []string `json:"rooms" bson:"rooms"`
}

func BuildRequests(api *gin.RouterGroup, api2 *gin.RouterGroup) {
	api.GET("", getScheduleOld)
	api.GET("view", getSchedule)
	api.PUT("", addLessons)
	api.GET("/types", getScheduleTypesOld)
	api.GET("/types/get", getTypes)

	api2.GET("/studyPlaces", getStudyPlaces)
}
