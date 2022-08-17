package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/entities"
	"studyum/src/utils"
	"time"
)

type ScheduleRepository struct {
	*Repository
}

func NewScheduleRepository(repository *Repository) *ScheduleRepository {
	return &ScheduleRepository{
		Repository: repository,
	}
}

func (s *ScheduleRepository) GetSchedule(ctx context.Context, studyPlaceId int, type_ string, typeName string, schedule *entities.Schedule) error {
	startWeekDate := utils.Date().AddDate(0, 0, 1-int(time.Now().Weekday()))
	cursor, err := s.studyPlacesCollection.Aggregate(ctx, bson.A{
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
		return err
	}

	cursor.Next(ctx)
	if err = cursor.Decode(&schedule); err != nil {
		return err
	}

	return nil
}

func (s *ScheduleRepository) GetScheduleType(ctx context.Context, studyPlaceId int, type_ string) []string {
	namesInterface, _ := s.lessonsCollection.Distinct(ctx, type_, bson.M{"studyPlaceId": studyPlaceId})

	names := make([]string, len(namesInterface))
	for i, v := range namesInterface {
		names[i] = v.(string)
	}

	return names
}

func (s *ScheduleRepository) AddLesson(ctx context.Context, lesson *entities.Lesson, studyPlaceId int) error {
	if lesson.Type == "GENERAL" {
		lesson.Type = "STAY"
	}

	lesson.Id = primitive.NewObjectID()
	lesson.StudyPlaceId = studyPlaceId
	if _, err := s.lessonsCollection.InsertOne(ctx, lesson); err != nil {
		return err
	}

	return nil
}

func (s *ScheduleRepository) UpdateLesson(ctx context.Context, lesson *entities.Lesson, studyPlaceId int) error {
	lesson.StudyPlaceId = studyPlaceId

	if _, err := s.lessonsCollection.UpdateOne(ctx, bson.M{"_id": lesson.Id, "studyPlaceId": studyPlaceId}, bson.M{"$set": lesson}); err != nil {
		return err
	}

	return nil
}

func (s *ScheduleRepository) DeleteLesson(ctx context.Context, id primitive.ObjectID, studyPlaceId int) error {
	if _, err := s.lessonsCollection.DeleteOne(ctx, bson.M{"_id": id, "studyPlaceId": studyPlaceId}); err != nil {
		return err
	}

	return nil
}
