package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
	"studyum/pkg/datetime"
	"studyum/pkg/slicetools"
	"time"
)

type ScheduleRepository interface {
	GetSchedule(ctx context.Context, studyPlaceId primitive.ObjectID, type_ string, typeName string, asPreview bool) (entities.Schedule, error)
	GetScheduleType(ctx context.Context, studyPlaceId primitive.ObjectID, type_ string) []string

	AddLessons(ctx context.Context, lessons []entities.Lesson) error
	AddGeneralLessons(ctx context.Context, lessons []entities.GeneralLesson) error
	AddLesson(ctx context.Context, lesson entities.Lesson) (primitive.ObjectID, error)
	UpdateLesson(ctx context.Context, lesson entities.Lesson) error
	FindAndDeleteLesson(ctx context.Context, id primitive.ObjectID, studyPlaceId primitive.ObjectID) (entities.Lesson, error)
	UpdateGeneralSchedule(ctx context.Context, lessons []entities.GeneralLesson) error
	RemoveLessonBetweenDates(ctx context.Context, date1, date2 time.Time, id primitive.ObjectID) error
	RemoveGroupLessonBetweenDates(ctx context.Context, date1, date2 time.Time, id primitive.ObjectID, group string) error

	RemoveGeneralLessonsByType(ctx context.Context, studyPlaceID primitive.ObjectID, type_ string, name string) error

	GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (err error, studyPlace entities.StudyPlace)
	GetGeneralLessons(ctx context.Context, studyPlaceId primitive.ObjectID, weekIndex, dayIndex int) ([]entities.GeneralLesson, error)
}

type scheduleRepository struct {
	*Repository
}

func NewScheduleRepository(repository *Repository) ScheduleRepository {
	return &scheduleRepository{Repository: repository}
}

func (s *scheduleRepository) GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (err error, studyPlace entities.StudyPlace) {
	err = s.studyPlacesCollection.FindOne(ctx, bson.M{"_id": id, "restricted": restricted}).Decode(&studyPlace)
	return
}

func (s *scheduleRepository) GetSchedule(ctx context.Context, studyPlaceID primitive.ObjectID, type_ string, typeName string, asPreview bool) (entities.Schedule, error) {
	filter := bson.M{"_id": studyPlaceID}
	if asPreview {
		filter["restricted"] = false
	}

	startWeekDate := datetime.Date().AddDate(0, 0, 1-int(time.Now().Weekday()))
	cursor, err := s.studyPlacesCollection.Aggregate(ctx, bson.A{
		bson.M{"$match": filter},
		bson.M{
			"$addFields": bson.M{
				"env": bson.M{
					"studyPlaceID": studyPlaceID,
					"startDate":    startWeekDate,
					"endDate": bson.M{
						"$dateAdd": bson.M{
							"startDate": startWeekDate,
							"unit":      "week",
							"amount":    "$weeksCount",
						},
					},
					"weeksAmount": "$weeksCount",
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "Lessons",
				"let":  bson.M{"env": "$env"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{
										"$eq": bson.A{"$studyPlaceId", "$$env.studyPlaceID"},
									}, bson.M{
										"$eq": bson.A{"$" + type_, typeName},
									}, bson.M{
										"$gte": bson.A{"$startDate", "$$env.startDate"},
									},
								},
							},
						},
					},
					bson.M{
						"$project": bson.M{
							//TODO user marks
							"marks":    0,
							"absences": 0,
						},
					},
					bson.M{
						"$addFields": bson.M{
							"isGeneral": false,
						},
					},
				},
				"as": "lessons",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"env.lastUpdatedDate": bson.M{"$max": "$lessons.endDate"},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"env.startGeneral": bson.M{
					"$dateFromParts": bson.M{
						"year":  bson.M{"$year": "$env.lastUpdatedDate"},
						"month": bson.M{"$month": "$env.lastUpdatedDate"},
						"day":   bson.M{"$sum": bson.A{bson.M{"$dayOfMonth": "$env.lastUpdatedDate"}, 1}},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"env.startWeekIndex":        bson.M{"$mod": bson.A{bson.M{"$isoWeek": "$env.startDate"}, "$env.weeksAmount"}},
				"env.startGeneralDayIndex":  bson.M{"$subtract": bson.A{bson.M{"$isoDayOfWeek": "$env.startGeneral"}, 1}},
				"env.startGeneralWeekIndex": bson.M{"$mod": bson.A{bson.M{"$isoWeek": "$env.startGeneral"}, "$env.weeksAmount"}},
				"env.endGeneralDayIndex":    bson.M{"$subtract": bson.A{bson.M{"$isoDayOfWeek": "$env.endDate"}, 1}},
				"env.endGeneralWeekIndex":   bson.M{"$mod": bson.A{bson.M{"$isoWeek": "$env.endDate"}, "$env.weeksAmount"}},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "GeneralLessons",
				"let":  bson.M{"env": "$env"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{
										"$eq": bson.A{"$studyPlaceId", "$$env.studyPlaceID"},
									}, bson.M{
										"$eq": bson.A{"$" + type_, typeName},
									},
								},
							},
						},
					},
					bson.M{
						"$addFields": bson.M{
							"date": bson.M{
								"$dateAdd": bson.M{
									"startDate": bson.M{
										"$dateAdd": bson.M{
											"startDate": "$$env.startDate",
											"unit":      "week",
											"amount":    bson.M{"$abs": bson.M{"$subtract": bson.A{"$weekIndex", "$$env.startWeekIndex"}}},
										},
									},
									"unit":   "day",
									"amount": "$dayIndex",
								},
							},
						},
					},
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$gte": bson.A{"$date", "$$env.startGeneral"}},
									bson.M{"$lt": bson.A{"$date", "$$env.endDate"}},
								},
							},
						},
					},
					bson.M{
						"$addFields": bson.M{
							"startDate": bson.M{
								"$toDate": bson.M{
									"$concat": bson.A{bson.M{
										"$dateToString": bson.M{
											"format": "%Y-%m-%d",
											"date":   "$date",
										},
									}, "T", "$startTime"},
								},
							},
							"endDate": bson.M{
								"$toDate": bson.M{
									"$concat": bson.A{bson.M{
										"$dateToString": bson.M{
											"format": "%Y-%m-%d",
											"date":   "$date",
										},
									}, "T", "$endTime"},
								},
							},
							"isGeneral": true,
						},
					},
				},
				"as": "general",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"lessons": bson.M{"$concatArrays": bson.A{"$lessons", "$general"}},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"_id": nil,
				"info": bson.M{
					"studyPlace":    "$$ROOT",
					"type":          type_,
					"typeName":      typeName,
					"startWeekDate": startWeekDate,
					"date":          time.Now(),
				},
				"lessons": "$lessons",
			},
		},
		bson.M{
			"$project": bson.M{
				"info.studyPlace.lessons": 0,
				"info.studyPlace.general": 0,
				"info.studyPlace.env":     0,
			},
		},
		bson.M{
			"$project": bson.M{
				"info":    1,
				"lessons": 1,
			},
		},
	})
	if err != nil {
		return entities.Schedule{}, err
	}

	if !cursor.Next(ctx) {
		var studyPlace entities.StudyPlace
		if err = s.studyPlacesCollection.FindOne(ctx, bson.M{"_id": studyPlaceID}).Decode(&studyPlace); err != nil {
			return entities.Schedule{}, err
		}

		return entities.Schedule{
			Info: entities.ScheduleInfo{
				Type:          type_,
				TypeName:      typeName,
				StudyPlace:    studyPlace,
				StartWeekDate: startWeekDate,
				Date:          time.Now(),
			},
		}, nil
	}

	var schedule entities.Schedule
	if err = cursor.Decode(&schedule); err != nil {
		return entities.Schedule{}, err
	}

	return schedule, nil
}

func (s *scheduleRepository) GetScheduleType(ctx context.Context, studyPlaceId primitive.ObjectID, type_ string) []string {
	namesInterface, _ := s.lessonsCollection.Distinct(ctx, type_, bson.M{"studyPlaceId": studyPlaceId})

	names := make([]string, len(namesInterface))
	for i, v := range namesInterface {
		names[i] = v.(string)
	}

	return names
}

func (s *scheduleRepository) AddGeneralLessons(ctx context.Context, lessons []entities.GeneralLesson) error {
	_, err := s.generalLessonsCollection.DeleteMany(ctx, bson.M{"studyPlaceId": lessons[0].StudyPlaceId})
	if err != nil {
		return err
	}

	_, err = s.generalLessonsCollection.InsertMany(ctx, slicetools.ToInterface(lessons))
	return err
}

func (s *scheduleRepository) AddLessons(ctx context.Context, lessons []entities.Lesson) error {
	_, err := s.lessonsCollection.InsertMany(ctx, slicetools.ToInterface(lessons))
	return err
}

func (s *scheduleRepository) AddLesson(ctx context.Context, lesson entities.Lesson) (primitive.ObjectID, error) {
	lesson.Id = primitive.NewObjectID()
	_, err := s.lessonsCollection.InsertOne(ctx, lesson)
	return lesson.Id, err
}

func (s *scheduleRepository) UpdateLesson(ctx context.Context, lesson entities.Lesson) error {
	_, err := s.lessonsCollection.UpdateOne(ctx, bson.M{"_id": lesson.Id, "studyPlaceId": lesson.StudyPlaceId}, bson.M{"$set": bson.M{
		"primaryColor":   lesson.PrimaryColor,
		"secondaryColor": lesson.PrimaryColor,
		"type":           lesson.Type,
		"endDate":        lesson.EndDate,
		"startDate":      lesson.StartDate,
		"subject":        lesson.Subject,
		"group":          lesson.Group,
		"teacher":        lesson.Teacher,
		"room":           lesson.Room,
		"title":          lesson.Title,
		"homework":       lesson.Homework,
		"description":    lesson.Description,
	}})
	return err
}

func (s *scheduleRepository) FindAndDeleteLesson(ctx context.Context, id primitive.ObjectID, studyPlaceId primitive.ObjectID) (entities.Lesson, error) {
	var lesson entities.Lesson
	err := s.lessonsCollection.FindOneAndDelete(ctx, bson.M{"_id": id, "studyPlaceId": studyPlaceId}).Decode(&lesson)
	return lesson, err
}

func (s *scheduleRepository) UpdateGeneralSchedule(ctx context.Context, lessons []entities.GeneralLesson) error {
	if _, err := s.generalLessonsCollection.InsertMany(ctx, slicetools.ToInterface(lessons)); err != nil {
		return err
	}

	return nil
}

func (s *scheduleRepository) GetGeneralLessons(ctx context.Context, studyPlaceId primitive.ObjectID, weekIndex, dayIndex int) ([]entities.GeneralLesson, error) {
	cursor, err := s.generalLessonsCollection.Find(ctx, bson.M{"studyPlaceId": studyPlaceId, "weekIndex": weekIndex, "dayIndex": dayIndex})
	if err != nil {
		return nil, err
	}

	var lessons []entities.GeneralLesson
	if err = cursor.All(ctx, &lessons); err != nil {
		return nil, err
	}

	return lessons, nil
}

func (s *scheduleRepository) RemoveLessonBetweenDates(ctx context.Context, date1, date2 time.Time, id primitive.ObjectID) error {
	_, err := s.lessonsCollection.DeleteMany(ctx, bson.M{"studyPlaceId": id, "startDate": bson.M{"$gte": date1, "$lt": date2}})
	return err
}

func (s *scheduleRepository) RemoveGroupLessonBetweenDates(ctx context.Context, date1, date2 time.Time, id primitive.ObjectID, group string) error {
	_, err := s.lessonsCollection.DeleteMany(ctx, bson.M{"studyPlaceId": id, "group": group, "startDate": bson.M{"$gte": date1, "$lt": date2}})
	return err
}

func (s *scheduleRepository) RemoveGeneralLessonsByType(ctx context.Context, studyPlaceID primitive.ObjectID, type_ string, typename string) error {
	_, err := s.generalLessonsCollection.DeleteMany(ctx, bson.M{"studyPlaceId": studyPlaceID, type_: typename})
	return err
}
