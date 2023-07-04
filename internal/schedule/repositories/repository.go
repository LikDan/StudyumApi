package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	general "studyum/internal/general/entities"
	"studyum/internal/schedule/entities"
	"studyum/pkg/hMongo"
	"studyum/pkg/slicetools"
	"time"
)

type Repository interface {
	GetSchedule(ctx context.Context, studyPlaceId primitive.ObjectID, role string, roleName string, startDate, endDate time.Time, general bool, asPreview bool) (entities.Schedule, error)
	GetScheduleType(ctx context.Context, studyPlaceId primitive.ObjectID, role string) []string

	AddGeneralLessons(ctx context.Context, lessons []entities.GeneralLesson) error

	AddLessons(ctx context.Context, lessons []entities.Lesson) error
	AddLesson(ctx context.Context, lesson entities.Lesson) error
	GetLessonByID(ctx context.Context, id primitive.ObjectID) (entities.Lesson, error)
	GetFullLessonByID(ctx context.Context, id primitive.ObjectID) (entities.Lesson, error)
	UpdateLesson(ctx context.Context, lesson entities.Lesson) error

	GetFullLessonsByIDAndDate(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) ([]entities.Lesson, error)

	DeleteLesson(ctx context.Context, id primitive.ObjectID, studyPlaceId primitive.ObjectID) error
	UpdateGeneralSchedule(ctx context.Context, lessons []entities.GeneralLesson) error
	RemoveLessonBetweenDates(ctx context.Context, date1, date2 time.Time, id primitive.ObjectID) error
	RemoveGroupLessonBetweenDates(ctx context.Context, date1, date2 time.Time, id primitive.ObjectID, group string) error

	RemoveGeneralLessonsByType(ctx context.Context, studyPlaceID primitive.ObjectID, role string, roleName string) error

	GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (err error, studyPlace general.StudyPlace)
	GetGeneralLessons(ctx context.Context, studyPlaceId primitive.ObjectID, weekIndex, dayIndex int) ([]entities.GeneralLesson, error)

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

func (s *repository) GetSchedule(ctx context.Context, studyPlaceID primitive.ObjectID, role string, roleName string, startDate, endDate time.Time, onlyGeneral bool, _ bool) (entities.Schedule, error) {
	cursor, err := s.generalLessons.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"studyPlaceId": studyPlaceID, role: roleName}},
		bson.M{"$group": bson.M{
			"_id":     bson.M{"dayIndex": "$dayIndex", "weekIndex": "$weekIndex"},
			"lessons": bson.M{"$push": "$$ROOT"},
		}},
		bson.M{
			"$sort": bson.M{
				"_id.weekIndex": 1,
				"_id.dayIndex":  1,
			},
		},
		bson.M{"$group": bson.M{
			"_id":  nil,
			"days": bson.M{"$push": "$$ROOT"},
		}},
		bson.M{
			"$addFields": bson.M{
				"days": bson.M{
					"$function": bson.M{
						"body": `function(templates, start, end) {
    						const getWeekNumber = (date) => {
								const yearStart = new Date(date.getFullYear(), 0, 1);
								return Math.ceil((((date.getTime() - yearStart.getTime()) / 86400000) + yearStart.getDay() + 1) / 7);
    						}

    						const weekAmount = Math.max(...templates.map(t => t._id.weekIndex)) + 1
							const currentDate = new Date(start.getTime());
							const lessons = [];
							while (currentDate <= end) {
								const day = currentDate.getUTCDay() === 0 ? 6 : currentDate.getUTCDay() - 1;
								const week = getWeekNumber(currentDate) % weekAmount;
                                const template = templates.find(t => t._id.dayIndex === day && t._id.weekIndex === week)
                                if (!!template) {
                                    template.lessons = template.lessons.map(t => {
                                        const date = new Date(currentDate.getTime())
                                        const startDate = new Date(currentDate.toLocaleDateString() + ' ' + t.startTime)
                                        const endDate = new Date(currentDate.toLocaleDateString() + ' ' + t.endTime)
                                        return {...t, date, startDate, endDate, isGeneral: true}
                                    })
                                    lessons.push({...template});
                                }
								currentDate.setDate(currentDate.getDate() + 1);
							}

							return lessons
						}`,
						"args": bson.A{"$days", startDate, endDate},
						"lang": "js",
					},
				},
			},
		},
		bson.M{"$unwind": "$days"},
		bson.M{"$replaceRoot": bson.M{"newRoot": "$days"}},
		bson.M{"$addFields": bson.M{"_id.date": bson.M{"$first": "$lessons.date"}}},
		bson.M{"$project": bson.M{"general": "$lessons"}},
		bson.M{
			"$lookup": bson.M{
				"from": "Lessons",
				"let":  bson.M{"from": "$_id.date", "till": bson.M{"$dateAdd": bson.M{"startDate": "$_id.date", "unit": "day", "amount": 1}}},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{onlyGeneral, false}},
									bson.M{"$eq": bson.A{"$" + role, roleName}},
									bson.M{"$eq": bson.A{"$studyPlaceId", studyPlaceID}},
									bson.M{"$gte": bson.A{"$startDate", "$$from"}},
									bson.M{"$lt": bson.A{"$startDate", "$$till"}},
								},
							},
						},
					},
					bson.M{
						"$sort": bson.M{
							"startDate": 1,
						},
					},
				},
				"as": "lessons",
			},
		},
		bson.M{
			"$project": bson.M{
				"lessons": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$eq": bson.A{"$lessons", bson.A{}}},
						"then": "$general",
						"else": "$lessons",
					},
				},
			},
		},
		bson.M{"$unwind": "$lessons"},
		bson.M{"$replaceRoot": bson.M{"newRoot": "$lessons"}},
		bson.M{
			"$group": bson.M{
				"_id": nil,
				"info": bson.M{"$first": bson.M{
					"studyPlaceID": studyPlaceID,
					"role":         role,
					"roleName":     roleName,
					"startDate":    startDate,
					"endDate":      endDate,
				}},
				"lessons": bson.M{"$push": "$$ROOT"},
			},
		},
	})
	if err != nil {
		return entities.Schedule{}, err
	}

	if !cursor.Next(ctx) {
		var studyPlace general.StudyPlace
		if err = s.studyPlaces.FindOne(ctx, bson.M{"_id": studyPlaceID}).Decode(&studyPlace); err != nil {
			return entities.Schedule{}, err
		}

		return entities.Schedule{
			Info: entities.Info{
				StudyPlaceID: primitive.NilObjectID,
				Role:         role,
				RoleName:     roleName,
				StartDate:    startDate,
				EndDate:      endDate,
				Date:         time.Now(),
			},
		}, nil
	}

	var schedule entities.Schedule
	if err = cursor.Decode(&schedule); err != nil {
		return entities.Schedule{}, err
	}

	return schedule, nil
}

func (s *repository) GetScheduleType(ctx context.Context, studyPlaceId primitive.ObjectID, role string) []string {
	namesInterface, _ := s.lessons.Distinct(ctx, role, bson.M{"studyPlaceId": studyPlaceId})

	names := make([]string, len(namesInterface))
	for i, v := range namesInterface {
		names[i] = v.(string)
	}

	return names
}

func (s *repository) AddGeneralLessons(ctx context.Context, lessons []entities.GeneralLesson) error {
	_, err := s.generalLessons.DeleteMany(ctx, bson.M{"studyPlaceId": lessons[0].StudyPlaceId})
	if err != nil {
		return err
	}

	_, err = s.generalLessons.InsertMany(ctx, slicetools.ToInterface(lessons))
	return err
}

func (s *repository) AddLessons(ctx context.Context, lessons []entities.Lesson) error {
	_, err := s.lessons.InsertMany(ctx, slicetools.ToInterface(lessons))
	return err
}

func (s *repository) AddLesson(ctx context.Context, lesson entities.Lesson) error {
	_, err := s.lessons.InsertOne(ctx, lesson)
	return err
}

func (s *repository) GetLessonByID(ctx context.Context, id primitive.ObjectID) (lesson entities.Lesson, err error) {
	opt := options.FindOne()
	opt.Projection = bson.M{"marks": 0, "absences": 0}

	err = s.lessons.FindOne(ctx, bson.M{"_id": id}, opt).Decode(&lesson)
	return
}

func (s *repository) GetFullLessonByID(ctx context.Context, id primitive.ObjectID) (lesson entities.Lesson, err error) {
	err = s.lessons.FindOne(ctx, bson.M{"_id": id}).Decode(&lesson)
	return
}

func (s *repository) UpdateLesson(ctx context.Context, lesson entities.Lesson) error {
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

func (s *repository) DeleteLesson(ctx context.Context, id primitive.ObjectID, studyPlaceId primitive.ObjectID) error {
	_, err := s.lessons.DeleteMany(ctx, bson.M{"_id": id, "studyPlaceId": studyPlaceId})
	return err
}

func (s *repository) UpdateGeneralSchedule(ctx context.Context, lessons []entities.GeneralLesson) error {
	if _, err := s.generalLessons.InsertMany(ctx, slicetools.ToInterface(lessons)); err != nil {
		return err
	}

	return nil
}

func (s *repository) GetGeneralLessons(ctx context.Context, studyPlaceId primitive.ObjectID, weekIndex, dayIndex int) ([]entities.GeneralLesson, error) {
	cursor, err := s.generalLessons.Find(ctx, bson.M{"studyPlaceId": studyPlaceId, "weekIndex": weekIndex, "dayIndex": dayIndex})
	if err != nil {
		return nil, err
	}

	var lessons []entities.GeneralLesson
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

func (s *repository) RemoveGeneralLessonsByType(ctx context.Context, studyPlaceID primitive.ObjectID, role string, roleName string) error {
	_, err := s.generalLessons.DeleteMany(ctx, bson.M{"studyPlaceId": studyPlaceID, role: roleName})
	return err
}

func (s *repository) FilterLessonMarks(ctx context.Context, lessonID primitive.ObjectID, marks []string) error {
	_, err := s.lessons.UpdateByID(ctx, lessonID, bson.M{"$pull": bson.M{"marks": bson.M{"mark": bson.M{"$nin": marks}}}})
	if err != nil && err.Error() == "write exception: write errors: [Cannot apply $pull to a non-array value]" {
		return nil
	}
	return err
}

func (s *repository) GetFullLessonsByIDAndDate(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) (lessons []entities.Lesson, err error) {
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
