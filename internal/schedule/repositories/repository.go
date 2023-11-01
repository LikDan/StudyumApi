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
	GetTypeID(ctx context.Context, studyPlaceId primitive.ObjectID, type_, typeName string) (primitive.ObjectID, error)

	GetSchedule(ctx context.Context, studyPlaceID primitive.ObjectID, type_, typeName string, typeID primitive.ObjectID, startDate, endDate time.Time, onlyGeneral bool, _ bool) (entities.Schedule, error)

	GetScheduleType(ctx context.Context, studyPlaceId primitive.ObjectID, role string, property string) (entries []entities.TypeEntry, err error)
	GetScheduleTeacherType(ctx context.Context, studyPlaceId primitive.ObjectID) (entries []entities.TypeEntry, err error)

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
	schedule       *mongo.Collection

	database *mongo.Database
}

func NewScheduleRepository(studyPlaces *mongo.Collection, lessons *mongo.Collection, generalLessons *mongo.Collection, schedule *mongo.Collection, database *mongo.Database) Repository {
	return &repository{studyPlaces: studyPlaces, lessons: lessons, generalLessons: generalLessons, schedule: schedule, database: database}
}

func (s *repository) GetTypeID(ctx context.Context, studyPlaceID primitive.ObjectID, type_, typeName string) (primitive.ObjectID, error) {
	var collection string
	switch type_ {
	case "teacher":
		var value struct {
			ID primitive.ObjectID `bson:"_id"`
		}

		err := s.database.Collection("Users").FindOne(ctx, bson.M{"studyPlaceInfo.roleName": typeName}).Decode(&value)
		if err != nil {
			return primitive.ObjectID{}, err
		}

		return value.ID, nil
	case "group", "student":
		type_ = "group"
		collection = "Groups"
	case "subject":
		collection = "Subjects"
	case "room":
		collection = "Rooms"
	}

	var value struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	err := s.database.Collection(collection).FindOne(ctx, bson.M{"studyPlaceID": studyPlaceID, type_: typeName}).Decode(&value)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return value.ID, nil
}

func (s *repository) GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (err error, studyPlace general.StudyPlace) {
	err = s.studyPlaces.FindOne(ctx, bson.M{"_id": id, "restricted": restricted}).Decode(&studyPlace)
	return
}

func (s *repository) GetSchedule(ctx context.Context, studyPlaceID primitive.ObjectID, type_, typeName string, typeID primitive.ObjectID, startDate, endDate time.Time, onlyGeneral bool, _ bool) (entities.Schedule, error) {
	if type_ == "student" {
		type_ = "group"
	}

	cursor, err := s.schedule.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"date":   bson.M{"$gte": startDate, "$lte": endDate},
				"status": bson.M{"$ne": "draft"},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "Lessons",
				"let":  bson.M{"date": "$date", "status": "$status"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$and": bson.A{
							bson.M{"$eq": bson.A{"$$date", bson.M{"$dateTrunc": bson.M{"date": "$startDate", "unit": "day"}}}},
							bson.M{"$eq": bson.A{"$" + type_ + "ID", typeID}},
						}}},
					},
					bson.M{
						"$addFields": bson.M{"status": "$$status"},
					},
				},
				"as": "lessons",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":   nil,
				"items": bson.M{"$push": "$$ROOT"}},
		},
		bson.M{
			"$project": bson.M{
				"items": bson.M{
					"$function": bson.M{
						"body": `function (items, start, end) { 
const currentDate = new Date(start);
while (currentDate <= end) {
    if (!items.some(item => item.date.getTime() === currentDate.getTime())) { 
        items.push({date: new Date(currentDate)}); 
    }
    currentDate.setDate(currentDate.getDate() + 1); 
}
return items; 
}`,
						"args": bson.A{"$items", startDate, endDate},
						"lang": "js",
					},
				},
			},
		},
		bson.M{"$unwind": "$items"},
		bson.M{"$replaceRoot": bson.M{"newRoot": "$items"}},
		bson.M{"$lookup": bson.M{
			"from": "GeneralLessons",
			"let":  bson.M{"lessons": "$lessons", "date": "$date"},
			"pipeline": bson.A{
				bson.M{
					"$match": bson.M{"$expr": bson.M{"$and": bson.A{
						bson.M{"$ne": bson.A{bson.M{"$type": "$$lessons"}, "array"}},
						bson.M{"$eq": bson.A{"$" + type_ + "ID", typeID}},
						bson.M{"$eq": bson.A{"$dayIndex", bson.M{"$dayOfWeek": "$$date"}}},
					}}},
				},
				bson.M{
					"$addFields": bson.M{
						"startDate": bson.M{"$dateAdd": bson.M{"startDate": "$$date", "unit": "minute", "amount": "$startTimeMinutes"}},
						"endDate":   bson.M{"$dateAdd": bson.M{"startDate": "$$date", "unit": "minute", "amount": "$endTimeMinutes"}},
						"status":    "general",
					},
				},
			}, "as": "generalLessons"},
		},
		bson.M{
			"$project": bson.M{
				"lessons": bson.M{"$cond": bson.M{
					"if":   bson.M{"$eq": bson.A{bson.M{"$type": "$lessons"}, "array"}},
					"then": "$lessons",
					"else": "$generalLessons"},
				}},
		},
		bson.M{"$unwind": "$lessons"},
		bson.M{"$replaceRoot": bson.M{"newRoot": "$lessons"}},
		bson.M{"$lookup": bson.M{"from": "StudyPlaceUsers", "localField": "teacherID", "foreignField": "_id", "as": "teacher"}},
		bson.M{"$lookup": bson.M{"from": "Groups", "localField": "groupID", "foreignField": "_id", "as": "group"}},
		bson.M{"$lookup": bson.M{"from": "Subjects", "localField": "subjectID", "foreignField": "_id", "as": "subject"}},
		bson.M{"$lookup": bson.M{"from": "Rooms", "localField": "roomID", "foreignField": "_id", "as": "room"}},
		bson.M{
			"$addFields": bson.M{
				"subject": bson.M{"$first": "$subject.subject"},
				"room":    bson.M{"$first": "$room.room"},
				"teacher": bson.M{"$first": "$teacher.roleName"},
				"group":   bson.M{"$first": "$group.group"},
			},
		},
		bson.M{"$project": bson.M{
			"endDate":        1,
			"startDate":      1,
			"subjectID":      1,
			"subject":        1,
			"teacherID":      1,
			"teacher":        1,
			"groupID":        1,
			"group":          1,
			"roomID":         1,
			"room":           1,
			"lessonIndex":    1,
			"primaryColor":   1,
			"secondaryColor": 1,
			"status":         1,
		}},
		bson.M{"$group": bson.M{
			"_id":     nil,
			"lessons": bson.M{"$push": "$$ROOT"},
			"info": bson.M{"$first": bson.M{
				"studyPlaceInfo": bson.M{"_id": studyPlaceID},
				"type":           type_,
				"typeName":       typeName, //todo remove
				"startDate":      startDate,
				"endDate":        endDate,
			}},
		}},
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
				StudyPlaceInfo: entities.StudyPlaceInfo{
					Id:    studyPlace.Id,
					Title: studyPlace.Name,
				},
				Type:      type_,
				TypeName:  "typeName",
				StartDate: startDate,
				EndDate:   endDate,
				Date:      time.Now(),
			},
		}, nil
	}

	var schedule entities.Schedule
	if err = cursor.Decode(&schedule); err != nil {
		return entities.Schedule{}, err
	}

	return schedule, nil
}

func (s *repository) GetScheduleType(ctx context.Context, studyPlaceId primitive.ObjectID, role string, property string) (entries []entities.TypeEntry, err error) {
	result, err := s.database.Collection(role).Find(ctx, bson.M{"studyPlaceID": studyPlaceId}, &options.FindOptions{
		Projection: bson.M{"_id": 1, "title": "$" + property},
	})
	if err != nil {
		return nil, err
	}

	err = result.All(ctx, &entries)
	return
}

func (s *repository) GetScheduleTeacherType(ctx context.Context, studyPlaceId primitive.ObjectID) (entries []entities.TypeEntry, err error) {
	result, err := s.database.Collection("StudyPlaceUsers").Find(ctx, bson.M{"studyPlaceID": studyPlaceId, "role": "teacher"}, &options.FindOptions{
		Projection: bson.M{"_id": 1, "title": "$roleName"},
	})
	if err != nil {
		return nil, err
	}

	err = result.All(ctx, &entries)
	return
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
