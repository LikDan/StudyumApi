package schedule

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	general "studyum/internal/general/entities"
	"studyum/pkg/datetime"
	"studyum/pkg/hMongo"
	"studyum/pkg/slicetools"
	"time"
)

type Repository interface {
	GetSchedule(ctx context.Context, studyPlaceId primitive.ObjectID, type_ string, typeName string, general bool, asPreview bool) (Schedule, error)
	GetScheduleType(ctx context.Context, studyPlaceId primitive.ObjectID, type_ string) []string

	AddGeneralLessons(ctx context.Context, lessons []GeneralLesson) error

	AddLessons(ctx context.Context, lessons []Lesson) error
	AddLesson(ctx context.Context, lesson Lesson) (primitive.ObjectID, error)
	GetLessonByID(ctx context.Context, id primitive.ObjectID) (Lesson, error)
	GetFullLessonByID(ctx context.Context, id primitive.ObjectID) (Lesson, error)
	UpdateLesson(ctx context.Context, lesson Lesson) error

	GetFullLessonsByIDAndDate(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) ([]Lesson, error)

	FindAndDeleteLesson(ctx context.Context, id primitive.ObjectID, studyPlaceId primitive.ObjectID) (Lesson, error)
	UpdateGeneralSchedule(ctx context.Context, lessons []GeneralLesson) error
	RemoveLessonBetweenDates(ctx context.Context, date1, date2 time.Time, id primitive.ObjectID) error
	RemoveGroupLessonBetweenDates(ctx context.Context, date1, date2 time.Time, id primitive.ObjectID, group string) error

	RemoveGeneralLessonsByType(ctx context.Context, studyPlaceID primitive.ObjectID, type_ string, name string) error

	GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (err error, studyPlace general.StudyPlace)
	GetGeneralLessons(ctx context.Context, studyPlaceId primitive.ObjectID, weekIndex, dayIndex int) ([]GeneralLesson, error)

	FilterLessonMarks(ctx context.Context, lessonID primitive.ObjectID, marks []string) error
}

type repository struct {
	studyPlaces    *mongo.Collection
	lessons        *mongo.Collection
	generalLessons *mongo.Collection
}

func NewScheduleRepository(studyPlaces *mongo.Collection, lessons *mongo.Collection, generalLessons *mongo.Collection) Repository {
	return &repository{studyPlaces: studyPlaces, lessons: lessons, generalLessons: generalLessons}
}

func (s *repository) GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (err error, studyPlace general.StudyPlace) {
	err = s.studyPlaces.FindOne(ctx, bson.M{"_id": id, "restricted": restricted}).Decode(&studyPlace)
	return
}

func (s *repository) GetSchedule(ctx context.Context, studyPlaceID primitive.ObjectID, type_ string, typeName string, isGeneral bool, asPreview bool) (Schedule, error) {
	filter := bson.M{"_id": studyPlaceID}
	if asPreview {
		filter["restricted"] = false
	}

	startWeekDate := datetime.Date().AddDate(0, 0, 1-int(time.Now().Weekday()))
	cursor, err := s.studyPlaces.Aggregate(ctx, bson.A{
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
										"$eq": bson.A{isGeneral, false},
									}, bson.M{
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
		return Schedule{}, err
	}

	if !cursor.Next(ctx) {
		var studyPlace general.StudyPlace
		if err = s.studyPlaces.FindOne(ctx, bson.M{"_id": studyPlaceID}).Decode(&studyPlace); err != nil {
			return Schedule{}, err
		}

		return Schedule{
			Info: Info{
				Type:          type_,
				TypeName:      typeName,
				StudyPlace:    studyPlace,
				StartWeekDate: startWeekDate,
				Date:          time.Now(),
			},
		}, nil
	}

	var schedule Schedule
	if err = cursor.Decode(&schedule); err != nil {
		return Schedule{}, err
	}

	return schedule, nil
}

func (s *repository) GetScheduleType(ctx context.Context, studyPlaceId primitive.ObjectID, type_ string) []string {
	namesInterface, _ := s.lessons.Distinct(ctx, type_, bson.M{"studyPlaceId": studyPlaceId})

	names := make([]string, len(namesInterface))
	for i, v := range namesInterface {
		names[i] = v.(string)
	}

	return names
}

func (s *repository) AddGeneralLessons(ctx context.Context, lessons []GeneralLesson) error {
	_, err := s.generalLessons.DeleteMany(ctx, bson.M{"studyPlaceId": lessons[0].StudyPlaceId})
	if err != nil {
		return err
	}

	_, err = s.generalLessons.InsertMany(ctx, slicetools.ToInterface(lessons))
	return err
}

func (s *repository) AddLessons(ctx context.Context, lessons []Lesson) error {
	_, err := s.lessons.InsertMany(ctx, slicetools.ToInterface(lessons))
	return err
}

func (s *repository) AddLesson(ctx context.Context, lesson Lesson) (primitive.ObjectID, error) {
	lesson.Id = primitive.NewObjectID()
	_, err := s.lessons.InsertOne(ctx, lesson)
	return lesson.Id, err
}

func (s *repository) GetLessonByID(ctx context.Context, id primitive.ObjectID) (lesson Lesson, err error) {
	opt := options.FindOne()
	opt.Projection = bson.M{"marks": 0, "absences": 0}

	err = s.lessons.FindOne(ctx, bson.M{"_id": id}, opt).Decode(&lesson)
	return
}

func (s *repository) GetFullLessonByID(ctx context.Context, id primitive.ObjectID) (lesson Lesson, err error) {
	err = s.lessons.FindOne(ctx, bson.M{"_id": id}).Decode(&lesson)
	return
}

func (s *repository) UpdateLesson(ctx context.Context, lesson Lesson) error {
	_, err := s.lessons.UpdateOne(ctx, bson.M{"_id": lesson.Id, "studyPlaceId": lesson.StudyPlaceId}, bson.M{"$set": bson.M{
		"primaryColor":   lesson.PrimaryColor,
		"secondaryColor": lesson.SecondaryColor,
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

func (s *repository) FindAndDeleteLesson(ctx context.Context, id primitive.ObjectID, studyPlaceId primitive.ObjectID) (Lesson, error) {
	var lesson Lesson
	err := s.lessons.FindOneAndDelete(ctx, bson.M{"_id": id, "studyPlaceId": studyPlaceId}).Decode(&lesson)
	return lesson, err
}

func (s *repository) UpdateGeneralSchedule(ctx context.Context, lessons []GeneralLesson) error {
	if _, err := s.generalLessons.InsertMany(ctx, slicetools.ToInterface(lessons)); err != nil {
		return err
	}

	return nil
}

func (s *repository) GetGeneralLessons(ctx context.Context, studyPlaceId primitive.ObjectID, weekIndex, dayIndex int) ([]GeneralLesson, error) {
	cursor, err := s.generalLessons.Find(ctx, bson.M{"studyPlaceId": studyPlaceId, "weekIndex": weekIndex, "dayIndex": dayIndex})
	if err != nil {
		return nil, err
	}

	var lessons []GeneralLesson
	if err = cursor.All(ctx, &lessons); err != nil {
		return nil, err
	}

	return lessons, nil
}

func (s *repository) RemoveLessonBetweenDates(ctx context.Context, date1, date2 time.Time, id primitive.ObjectID) error {
	_, err := s.lessons.DeleteMany(ctx, bson.M{"studyPlaceId": id, "startDate": bson.M{"$gte": date1, "$lt": date2}})
	return err
}

func (s *repository) RemoveGroupLessonBetweenDates(ctx context.Context, date1, date2 time.Time, id primitive.ObjectID, group string) error {
	_, err := s.lessons.DeleteMany(ctx, bson.M{"studyPlaceId": id, "group": group, "startDate": bson.M{"$gte": date1, "$lt": date2}})
	return err
}

func (s *repository) RemoveGeneralLessonsByType(ctx context.Context, studyPlaceID primitive.ObjectID, type_ string, typename string) error {
	_, err := s.generalLessons.DeleteMany(ctx, bson.M{"studyPlaceId": studyPlaceID, type_: typename})
	return err
}

func (s *repository) FilterLessonMarks(ctx context.Context, lessonID primitive.ObjectID, marks []string) error {
	_, err := s.lessons.UpdateByID(ctx, lessonID, bson.M{"$pull": bson.M{"marks": bson.M{"mark": bson.M{"$nin": marks}}}})
	if err != nil && err.Error() == "write exception: write errors: [Cannot apply $pull to a non-array value]" {
		return nil
	}
	return err
}

func (s *repository) GetFullLessonsByIDAndDate(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) (lessons []Lesson, err error) {
	cursor, err := s.lessons.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{"_id": id},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "Lessons",
				"let": bson.M{
					"date": bson.M{"$dateToString": bson.M{
						"date":   "$startDate",
						"format": "%Y-%m-%d",
					}},
					"group":   "$group",
					"teacher": "$teacher",
					"subject": "$subject",
				},
				"pipeline": bson.A{
					bson.M{"$match": bson.M{"$expr": bson.M{"$and": bson.A{
						bson.M{"$eq": bson.A{
							"$$date", bson.M{"$dateToString": bson.M{
								"date":   "$startDate",
								"format": "%Y-%m-%d",
							}},
						}},
						bson.M{"$eq": bson.A{"$group", "$$group"}},
						bson.M{"$eq": bson.A{"$subject", "$$subject"}},
						bson.M{"$eq": bson.A{"$teacher", "$$teacher"}},
					}}}},
					bson.M{
						"$addFields": bson.M{
							"marks":    hMongo.Filter("marks", bson.M{"$eq": bson.A{"$$marks.studentID", userID}}),
							"absences": hMongo.Filter("absences", bson.M{"$eq": bson.A{"$$absences.studentID", userID}}),
						},
					},
				},
				"as": "lessons",
			},
		},
		bson.M{
			"$project": bson.M{
				"lessons": 1,
			},
		},
		bson.M{
			"$unwind": "$lessons",
		},
		bson.M{
			"$replaceRoot": bson.M{"newRoot": "$lessons"},
		},
		bson.M{
			"$sort": bson.M{
				"startDate": 1,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &lessons)
	return
}
