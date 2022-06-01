package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	h "studyum/src/api"
	"studyum/src/models"
	"time"
)

func GetSchedule(studyPlaceId int, type_ string, typeName string, schedule *models.Schedule) *models.Error {
	startWeekDate := h.Date().AddDate(0, 0, 1-int(time.Now().Weekday()))
	cursor, err := StudyPlacesCollection.Aggregate(nil, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": studyPlaceId,
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
				"from": "GeneralLessons",
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
							"type":      "GENERAL",
							"startDate": bson.M{"$toDate": bson.M{"$concat": bson.A{bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$$date"}}, "T", "$startTime"}}},
							"endDate":   bson.M{"$toDate": bson.M{"$concat": bson.A{bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$$date"}}, "T", "$endTime"}}},
						},
					},
				},
				"as": "general",
			},
		}, bson.M{
			"$lookup": bson.M{
				"from": "Lessons",
				"let":  bson.M{"date": "$date"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$startDate"}}, bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$$date"}}}},
									bson.M{"$eq": bson.A{"$" + type_, typeName}},
								},
							},
						},
					}, bson.M{
						"$addFields": bson.M{
							"startDate": "$startDate",
							"endDate":   "$endDate",
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
	if err != nil {
		return models.BindError(err, 403, h.UNDEFINED)
	}

	cursor.Next(nil)
	if err = cursor.Decode(&schedule); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}

func GetScheduleType(studyPlaceId int, type_ string) []string {
	namesInterface, _ := LessonsCollection.Distinct(nil, type_, bson.M{"studyPlaceId": studyPlaceId})

	names := make([]string, len(namesInterface))
	for i, v := range namesInterface {
		names[i] = v.(string)
	}

	return names
}

func AddLesson(lesson *models.Lesson, studyPlaceId int) *models.Error {
	if lesson.Type == "GENERAL" {
		lesson.Type = "STAY"
	}

	lesson.Id = primitive.NewObjectID()
	lesson.StudyPlaceId = studyPlaceId
	if _, err := LessonsCollection.InsertOne(nil, lesson); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}

func UpdateLesson(lesson *models.Lesson, studyPlaceId int) *models.Error {
	lesson.StudyPlaceId = studyPlaceId

	if _, err := LessonsCollection.UpdateOne(nil, bson.M{"_id": lesson.Id, "studyPlaceId": studyPlaceId}, bson.M{"$set": lesson}); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}

func DeleteLesson(id primitive.ObjectID, studyPlaceId int) *models.Error {
	if _, err := LessonsCollection.DeleteOne(nil, bson.M{"_id": id, "studyPlaceId": studyPlaceId}); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}
