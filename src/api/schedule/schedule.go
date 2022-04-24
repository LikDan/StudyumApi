package schedule

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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
	studyPlaceIdStr := ctx.Query("studyPlaceId")

	var studyPlaceId int
	if studyPlaceIdStr == "" {
		studyPlaceId = user.StudyPlaceId
	} else {
		var err error
		studyPlaceId, err = strconv.Atoi(studyPlaceIdStr)
		if h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
			return
		}
	}

	if !h.CheckNotEmpty(type_, typeName) {
		h.ErrorMessage(ctx, "Provide valid params")
		return
	}

	var schedule Schedule

	startWeekDate := h.Date().AddDate(0, 0, 1-int(time.Now().Weekday()))
	cursor, err := db.GeneralSubjectsCollection.Aggregate(nil, bson.A{
		bson.M{
			"$match": bson.M{
				type_:              typeName,
				"educationPlaceId": studyPlaceId,
			},
		}, bson.M{
			"$group": bson.M{
				"_id":     "$weekIndex",
				"lessons": bson.M{"$push": "$$ROOT"},
			},
		}, bson.M{
			"$group": bson.M{
				"_id":           nil,
				"weeks":         bson.M{"$push": "$$ROOT"},
				"weeksQuantity": bson.M{"$sum": 1},
			},
		}, bson.M{
			"$unwind": "$weeks",
		}, bson.M{
			"$sort": bson.M{"weeks._id": 1},
		}, bson.M{
			"$group": bson.M{
				"_id": nil,
				"start": bson.M{"$push": bson.M{"$cond": bson.A{
					bson.M{"$gte": bson.A{
						bson.M{"$mod": bson.A{bson.M{"$isoWeek": startWeekDate}, "$weeksQuantity"}},
						"$weeks_.id",
					}},
					"$weeks",
					"$$REMOVE",
				}}},
				"end": bson.M{"$push": bson.M{"$cond": bson.A{
					bson.M{"$lt": bson.A{
						bson.M{"$mod": bson.A{bson.M{"$isoWeek": startWeekDate}, "$weeksQuantity"}},
						"$weeks_.id",
					}},
					"$weeks",
					"$$REMOVE",
				}}},
			},
		}, bson.M{
			"$project": bson.M{"weeks": bson.M{"$concatArrays": bson.A{"$start", "$end"}}},
		}, bson.M{
			"$unwind": bson.M{
				"path":              "$weeks",
				"includeArrayIndex": "index",
			},
		}, bson.M{
			"$project": bson.M{
				"lessons": "$weeks.lessons",
				"startWeekDate": bson.M{"$dateAdd": bson.M{
					"startDate": startWeekDate,
					"unit":      "week",
					"amount":    "$index",
				}},
			},
		}, bson.M{
			"$unwind": "$lessons",
		}, bson.M{
			"$addFields": bson.M{
				"lessons.startTime": bson.M{"$dateFromParts": bson.M{
					"year":   bson.M{"$year": "$startWeekDate"},
					"month":  bson.M{"$month": "$startWeekDate"},
					"day":    bson.M{"$sum": bson.A{bson.M{"$dayOfMonth": "$startWeekDate"}, "$lessons.columnIndex"}},
					"hour":   bson.M{"$hour": "$lessons.startTime"},
					"minute": bson.M{"$minute": "$lessons.startTime"},
				}},
				"lessons.endTime": bson.M{"$dateFromParts": bson.M{
					"year":   bson.M{"$year": "$startWeekDate"},
					"month":  bson.M{"$month": "$startWeekDate"},
					"day":    bson.M{"$sum": bson.A{bson.M{"$dayOfMonth": "$startWeekDate"}, "$lessons.columnIndex"}},
					"hour":   bson.M{"$hour": "$lessons.endTime"},
					"minute": bson.M{"$minute": "$lessons.endTime"},
				}},
				"lessons.updated": false,
			},
		}, bson.M{
			"$group": bson.M{
				"_id":     bson.M{"$dateToString": bson.M{"date": "$lessons.startTime", "format": "%Y-%m-%d"}},
				"general": bson.M{"$push": "$lessons"},
			},
		}, bson.M{
			"$lookup": bson.M{
				"from": "Subjects",
				"let":  bson.M{"date": "$_id"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$$date", bson.M{"$dateToString": bson.M{"date": "$startTime", "format": "%Y-%m-%d"}}}},
									bson.M{"$eq": bson.A{"$group", typeName}},
									bson.M{"$eq": bson.A{"$educationPlaceId", studyPlaceId}},
								},
							},
						},
					}, bson.M{
						"$addFields": bson.M{
							"updated": true,
						},
					},
				},
				"as": "current",
			},
		}, bson.M{
			"$project": bson.M{
				"lessons": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$current", bson.A{}}}, "$general", "$current"}},
			},
		}, bson.M{
			"$unwind": "$lessons",
		}, bson.M{
			"$replaceRoot": bson.M{"newRoot": "$lessons"},
		}, bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"startDate": "$startTime",
					"endDate":   "$endTime",
				},
				"startDate": bson.M{"$first": "$startTime"},
				"endDate":   bson.M{"$first": "$endTime"},
				"updated":   bson.M{"$first": "$updated"},
				"subjects": bson.M{"$push": bson.M{
					"subject":     "$subject",
					"teacher":     "$teacher",
					"group":       "$group",
					"room":        "$room",
					"type":        "$type",
					"title":       "$smalldescription",
					"description": "$description",
					"homework":    "$homework",
				}},
				"studyPlaceId": bson.M{"$first": "$educationPlaceId"},
			},
		}, bson.M{
			"$project": bson.M{"_id": 0},
		}, bson.M{
			"$sort": bson.M{"_id.startDate": 1},
		}, bson.M{
			"$group": bson.M{
				"_id":          nil,
				"lessons":      bson.M{"$push": "$$ROOT"},
				"studyPlaceId": bson.M{"$first": "$studyPlaceId"},
			},
		}, bson.M{
			"$lookup": bson.M{
				"from": "StudyPlaces",
				"let":  bson.M{"studyPlaceId": "$studyPlaceId"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$eq": bson.A{"$$studyPlaceId", "$_id"},
							},
						},
					},
				},
				"as": "studyPlace",
			},
		}, bson.M{
			"$project": bson.M{
				"lessons":         1,
				"info.studyPlace": bson.M{"$first": "$studyPlace"},
			},
		}, bson.M{
			"$addFields": bson.M{
				"info.type":          type_,
				"info.typeName":      typeName,
				"info.startWeekDate": startWeekDate,
				"info.date":          time.Now(),
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
	studyPlaceIdStr := ctx.Query("studyPlaceId")
	if studyPlaceIdStr == "" {
		h.ErrorMessage(ctx, "provide all params")
		return
	}

	studyPlaceId, err := strconv.Atoi(studyPlaceIdStr)
	if h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	var get = func(type_ string) []string {
		namesInterface, _ := db.SubjectsCollection.Distinct(nil, type_, bson.M{"educationPlaceId": studyPlaceId})

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
	api.GET("/types", getScheduleTypesOld)
	api.GET("/types/get", getTypes)

	api2.GET("/studyPlaces", getStudyPlaces)
}
