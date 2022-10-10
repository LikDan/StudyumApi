package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/entities"
)

type JournalRepository interface {
	AddMark(ctx context.Context, mark entities.Mark) (primitive.ObjectID, error)
	UpdateMark(ctx context.Context, mark entities.Mark) error
	GetMarkById(ctx context.Context, id primitive.ObjectID) (entities.Mark, error)
	DeleteMarkByID(ctx context.Context, id primitive.ObjectID) error

	GetAvailableOptions(ctx context.Context, teacher string, editable bool) ([]entities.JournalAvailableOption, error)

	GetStudentJournal(ctx context.Context, userId primitive.ObjectID, group string, studyPlaceId primitive.ObjectID) (entities.Journal, error)
	GetJournal(ctx context.Context, group string, subject string, typeName string, studyPlaceId primitive.ObjectID) (entities.Journal, error)
	GetAbsentJournal(ctx context.Context, group string, subject string, name string, id primitive.ObjectID) (entities.Journal, error)

	GetLessonByID(ctx context.Context, id primitive.ObjectID) (entities.Lesson, error)
	GetLessons(ctx context.Context, userId primitive.ObjectID, group, teacher, subject string, studyPlaceId primitive.ObjectID) ([]entities.Lesson, error)

	AddAbsence(ctx context.Context, absences entities.Absences) error
	UpdateAbsence(ctx context.Context, absences entities.Absences) error
	GetAbsenceByID(ctx context.Context, id primitive.ObjectID) (entities.Absences, error)
	DeleteAbsenceByID(ctx context.Context, id primitive.ObjectID) error
}

type journalRepository struct {
	*Repository
}

func NewJournalRepository(repository *Repository) JournalRepository {
	return &journalRepository{Repository: repository}
}

func (j *journalRepository) AddMark(ctx context.Context, mark entities.Mark) (primitive.ObjectID, error) {
	mark.Id = primitive.NewObjectID()
	if _, err := j.marksCollection.InsertOne(ctx, mark); err != nil {
		return primitive.NilObjectID, err
	}

	return mark.Id, nil
}

func (j *journalRepository) UpdateMark(ctx context.Context, mark entities.Mark) error {
	_, err := j.marksCollection.UpdateOne(ctx, bson.M{"_id": mark.Id, "lessonId": mark.LessonID}, bson.M{"$set": bson.M{"mark": mark.Mark}})
	return err
}

func (j *journalRepository) GetMarkById(ctx context.Context, id primitive.ObjectID) (entities.Mark, error) {
	var mark entities.Mark
	err := j.marksCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&mark)
	return mark, err
}

func (j *journalRepository) DeleteMarkByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := j.marksCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (j *journalRepository) GetAvailableOptions(ctx context.Context, teacher string, editable bool) ([]entities.JournalAvailableOption, error) {
	aggregate, err := j.lessonsCollection.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"teacher": teacher}},
		bson.M{"$group": bson.M{
			"_id": bson.M{
				"teacher": "$teacher",
				"subject": "$subject",
				"group":   "$group",
			},
			"teacher": bson.M{"$first": "$teacher"},
			"subject": bson.M{"$first": "$subject"},
			"group":   bson.M{"$first": "$group"}},
		},
		bson.M{"$addFields": bson.M{"editable": editable}},
		bson.M{"$sort": bson.M{"group": 1, "subject": 1, "teacher": 1}},
	})
	if err != nil {
		return nil, err
	}

	var options []entities.JournalAvailableOption
	if err = aggregate.All(ctx, &options); err != nil {
		return nil, err
	}

	return options, nil
}

func (j *journalRepository) GetStudentJournal(ctx context.Context, userId primitive.ObjectID, group string, studyPlaceId primitive.ObjectID) (entities.Journal, error) {
	cursor, err := j.lessonsCollection.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"group": group, "studyPlaceId": studyPlaceId}},
		bson.M{
			"$lookup": bson.M{
				"from":         "Marks",
				"localField":   "_id",
				"foreignField": "lessonID",
				"pipeline":     bson.A{bson.M{"$match": bson.M{"studentID": userId, "studyPlaceID": studyPlaceId}}},
				"as":           "marks",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":     "StudyPlaces",
				"pipeline": bson.A{bson.M{"$match": bson.M{"_id": studyPlaceId}}},
				"as":       "studyPlace",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"journalCellColor": bson.M{
					"$function": bson.M{
						// language=JavaScript
						"body": `function (studyPlace, lesson) {
                        if (lesson === undefined || lesson.marks === undefined) return "";

                        let color = studyPlace.journalColors.general
                        for (let mark of lesson.marks) {
                            let type = studyPlace.lessonTypes.find(v => v.type === lesson.type);
                            if (type === undefined) return studyPlace.journalColors.general;

                            let markType = type.marks.find(m => m.mark === mark.mark);
                            if (markType === undefined || markType.workOutTime === undefined) return studyPlace.journalColors.general;

                            lesson.startDate.setSeconds(lesson.startDate.getSeconds() + 604800);

                            color = lesson.startDate.getTime() > new Date().getTime() ? studyPlace.journalColors.warning : studyPlace.journalColors.danger;
                        }

                        return color;
                    }`,
						"args": bson.A{bson.M{"$first": "$studyPlace"}, "$$ROOT"},
						"lang": "js",
					},
				},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":        nil,
				"studyPlace": bson.M{"$first": bson.M{"$first": "$studyPlace"}},
				"lessons":    bson.M{"$push": "$$ROOT"},
				"dates":      bson.M{"$addToSet": bson.M{"$toDate": bson.M{"$dateToString": bson.M{"date": "$startDate", "format": "%m/%d/%Y"}}}},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"dates": bson.M{
					"$sortArray": bson.M{
						"input":  "$dates",
						"sortBy": 1,
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"lessons.studyPlace": 0,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"rows": bson.M{
					"$function": bson.M{
						// language=JavaScript
						"body": `function (studyPlace, lessons, dates) {
                        const groupBy = function (xs, key) {
                            return xs.reduce(function (rv, x) {
                                //TODO sort
                                (rv[x[key]] = rv[x[key]] || []).push(x);
                                return rv;
                            }, {});
                        };

                        let groupedLessons = groupBy(lessons, 'subject')

                        for (const [key, value] of Object.entries(groupedLessons)) {
                            groupedLessons[key] = value.sort((a, b) => a.startDate - b.startDate)
                        }

                        let rows = []
                        for (const [key, value] of Object.entries(groupedLessons)) {
                            rows.unshift({title: key, lessons: []})

                            let added = 0
                            for (let i = 0; i < value.length; i++) {
                                let startTime = new Date(value[i].startDate.toDateString()).getTime()
                                if (i > 0 && new Date(value[i - 1].startDate.toDateString()).getTime() === startTime) {
                                    let prevLesson = rows[0].lessons.at(-1)

                                    if (value[i].journalCellColor != studyPlace.journalColors.general && prevLesson.journalCellColor == studyPlace.journalColors.general) {
                                        prevLesson.journalCellColor = value[i].journalCellColor
                                    }

                                    prevLesson.marks = prevLesson.marks.concat(value[i].marks)
                                    added--
                                    continue
                                }
                                while (dates[i + added].getTime() !== startTime) {
                                    rows[0].lessons.push(null)
                                    added++
                                }
                                rows[0].lessons.push(value[i])
                            }
                            for (let i = added + value.length; i < dates.length; i++) {
                                rows[0].lessons.push(null)
                            }
                        }

                        return rows
                    }`,
						"args": bson.A{"$studyPlace", "$lessons", "$dates"},
						"lang": "js",
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"info": bson.M{
					"editable":   false,
					"studyPlace": "$studyPlace",
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"dates": bson.M{
					"$map": bson.M{
						"input": "$dates",
						"as":    "date",
						"in": bson.M{
							"startDate": "$$date",
							"endDate":   "$$date",
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"lessons":    0,
				"studyPlace": 0,
			},
		},
	})
	if err != nil {
		return entities.Journal{}, err
	}

	if !cursor.Next(ctx) {
		return entities.Journal{
			Info: entities.JournalInfo{
				Editable:   false,
				StudyPlace: entities.StudyPlace{},
				Group:      group,
			},
		}, nil
	}
	var journal entities.Journal
	if err = cursor.Decode(&journal); err != nil {
		return entities.Journal{}, err
	}

	return journal, nil
}

func (j *journalRepository) GetJournal(ctx context.Context, group string, subject string, typeName string, studyPlaceId primitive.ObjectID) (entities.Journal, error) {
	var cursor, err = j.usersCollection.Aggregate(ctx, bson.A{
		bson.M{"$group": bson.M{"_id": nil, "users": bson.M{"$push": "$$ROOT"}}},
		bson.M{"$lookup": bson.M{
			"from":     "SignUpCodes",
			"pipeline": mongo.Pipeline{},
			"as":       "codeUsers",
		}},
		bson.M{"$project": bson.M{"_id": nil, "users": bson.M{"$concatArrays": bson.A{"$users", "$codeUsers"}}}},
		bson.M{"$unwind": "$users"},
		bson.M{"$replaceRoot": bson.M{"newRoot": "$users"}},
		bson.M{"$match": bson.M{"type": "group", "typename": group, "studyPlaceID": studyPlaceId}},
		bson.M{"$lookup": bson.M{
			"from":     "Lessons",
			"pipeline": mongo.Pipeline{bson.D{{"$match", bson.M{"subject": subject, "teacher": typeName, "group": group, "studyPlaceId": studyPlaceId}}}},
			"as":       "subjects",
		}},
		bson.M{"$unwind": "$subjects"},
		bson.M{"$lookup": bson.M{
			"from": "StudyPlaces",
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"_id": studyPlaceId}},
			},
			"as": "studyPlace",
		}},
		bson.M{"$lookup": bson.M{
			"from":         "Marks",
			"localField":   "subjects._id",
			"foreignField": "lessonID",
			"let":          bson.M{"studentID": "$_id"},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$studentID", "$$studentID"}}}},
			},
			"as": "subjects.marks",
		}},
		bson.M{"$addFields": bson.M{
			"userType":           "student",
			"subjects.studentID": "$_id",
			"subjects.journalCellColor": bson.M{"$function": bson.M{
				"body": `function(studyPlace_, lesson) {
	let studyPlace = studyPlace_[0];
    
    color = ""
    for (let mark of lesson.marks) {
        let type = studyPlace.lessonTypes.find(v => v.type === lesson.type);
		if (type === undefined) return "white";
		
		let markType = type.marks.find(m => m.mark === mark.mark);
		if (markType === undefined || markType.workOutTime === undefined) return "white";
        
        let currentDate = new Date();
        let lessonDate = lesson.startDate;
        lessonDate.setSeconds(lessonDate.getSeconds() + 604800);
        
        color = lessonDate.getTime() > currentDate.getTime() ? "orange" : "red";
    }
    
    return color;
}`,
				"args": bson.A{"$studyPlace", "$subjects"},
				"lang": "js",
			}},
		}},
		bson.M{"$sort": bson.M{"subjects.startDate": 1}},
		bson.M{"$group": bson.M{"_id": "$_id", "title": bson.M{"$first": "$name"}, "userType": bson.M{"$first": "$userType"}, "lessons": bson.M{"$push": "$subjects"}, "studyPlace": bson.M{"$first": "$studyPlace"}}},
		bson.M{"$group": bson.M{"_id": nil, "rows": bson.M{"$push": "$$ROOT"}, "studyPlace": bson.M{"$first": "$studyPlace"}}},
		bson.M{"$project": bson.M{"_id": 0}},
		bson.M{"$lookup": bson.M{
			"from":     "Lessons",
			"pipeline": mongo.Pipeline{bson.D{{"$match", bson.M{"subject": subject, "teacher": typeName, "group": group, "studyPlaceId": studyPlaceId}}}, bson.D{{"$sort", bson.M{"startDate": 1}}}},
			"as":       "dates",
		}},
		bson.M{"$addFields": bson.M{"info": bson.M{
			"editable":   true,
			"studyPlace": bson.M{"$first": "$studyPlace"},
			"group":      group,
			"teacher":    typeName,
			"subject":    subject,
		}}},
	})
	if err != nil {
		return entities.Journal{}, err
	}

	if !cursor.Next(ctx) {
		return entities.Journal{
			Info: entities.JournalInfo{
				Editable:   true,
				StudyPlace: entities.StudyPlace{},
				Group:      group,
				Teacher:    typeName,
				Subject:    subject,
			},
		}, nil
	}
	var journal entities.Journal
	if err = cursor.Decode(&journal); err != nil {
		return entities.Journal{}, err
	}

	return journal, nil
}

func (j *journalRepository) GetAbsentJournal(ctx context.Context, group string, subject string, typeName string, studyPlaceId primitive.ObjectID) (entities.Journal, error) {
	var studyPlace entities.StudyPlace
	if err := j.studyPlacesCollection.FindOne(ctx, bson.M{"_id": studyPlaceId}).Decode(&studyPlace); err != nil {
		return entities.Journal{}, err
	}

	cursor, err := j.usersCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$group", bson.M{"_id": nil, "users": bson.M{"$push": "$$ROOT"}}}},
		bson.D{{"$lookup", bson.M{
			"from":     "SignUpCodes",
			"pipeline": mongo.Pipeline{},
			"as":       "codeUsers",
		}}},
		bson.D{{"$project", bson.M{"_id": nil, "users": bson.M{"$concatArrays": bson.A{"$users", "$codeUsers"}}}}},
		bson.D{{"$unwind", "$users"}},
		bson.D{{"$replaceRoot", bson.M{"newRoot": "$users"}}},
		bson.D{{"$match", bson.M{"type": "group", "typename": group, "studyPlaceID": studyPlaceId}}},
		bson.D{{"$lookup", bson.M{
			"from":     "Lessons",
			"pipeline": mongo.Pipeline{bson.D{{"$match", bson.M{"subject": subject, "teacher": typeName, "group": group, "studyPlaceId": studyPlaceId}}}},
			"as":       "subjects",
		}}},
		bson.D{{"$unwind", "$subjects"}},
		bson.D{{"$lookup", bson.M{
			"from":         "Absences",
			"localField":   "subjects._id",
			"foreignField": "lessonID",
			"let":          bson.M{"studentID": "$_id"},
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"$expr": bson.M{"$eq": bson.A{"$studentID", "$$studentID"}}}}},
				bson.D{{"$addFields", bson.M{"mark": bson.M{"$convert": bson.M{
					"input":  "$time",
					"to":     "string",
					"onNull": studyPlace.AbsentMark,
				}}}}},
			},
			"as": "subjects.marks",
		}}},
		bson.D{{"$sort", bson.M{"subjects.startDate": 1}}},
		bson.D{{"$addFields", bson.M{"userType": "student", "subjects.studentID": "$_id"}}},
		bson.D{{"$group", bson.M{"_id": "$_id", "title": bson.M{"$first": "$name"}, "userType": bson.M{"$first": "$userType"}, "lessons": bson.M{"$push": "$subjects"}}}},
		bson.D{{"$group", bson.M{"_id": nil, "rows": bson.M{"$push": "$$ROOT"}}}},
		bson.D{{"$project", bson.M{"_id": 0}}},
		bson.D{{"$lookup", bson.M{
			"from":     "Lessons",
			"pipeline": mongo.Pipeline{bson.D{{"$match", bson.M{"subject": subject, "teacher": typeName, "group": group, "studyPlaceId": studyPlaceId}}}, bson.D{{"$sort", bson.M{"startDate": 1}}}},
			"as":       "dates",
		}}},
		bson.D{{"$addFields", bson.M{"info": bson.M{
			"editable":   true,
			"studyPlace": studyPlace,
			"group":      group,
			"teacher":    typeName,
			"subject":    subject,
		}}}},
	})
	if err != nil {
		return entities.Journal{}, err
	}

	if !cursor.Next(ctx) {
		return entities.Journal{
			Info: entities.JournalInfo{
				Editable:   true,
				StudyPlace: studyPlace,
				Group:      group,
				Teacher:    typeName,
				Subject:    subject,
			},
		}, nil
	}
	var journal entities.Journal
	if err = cursor.Decode(&journal); err != nil {
		return entities.Journal{}, err
	}

	return journal, nil
}

func (j *journalRepository) GetLessonByID(ctx context.Context, id primitive.ObjectID) (lesson entities.Lesson, err error) {
	err = j.lessonsCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&lesson)
	return
}

func (j *journalRepository) GetLessons(ctx context.Context, userId primitive.ObjectID, group, teacher, subject string, studyPlaceId primitive.ObjectID) ([]entities.Lesson, error) {
	lessonsCursor, err := j.lessonsCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$lookup", bson.M{
			"from":         "Marks",
			"localField":   "_id",
			"foreignField": "lessonId",
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"studentID": userId}}},
			},
			"as": "marks",
		}}},
		bson.D{{"$match", bson.M{"group": group, "teacher": teacher, "subject": subject, "studyPlaceId": studyPlaceId}}},
		bson.D{{"$sort", bson.M{"date": 1}}},
	})

	var marks []entities.Lesson
	if err = lessonsCursor.All(ctx, &marks); err != nil {
		return nil, err
	}

	return marks, nil
}

func (j *journalRepository) AddAbsence(ctx context.Context, absences entities.Absences) error {
	_, err := j.absencesCollection.InsertOne(ctx, absences)
	return err
}

func (j *journalRepository) UpdateAbsence(ctx context.Context, absences entities.Absences) error {
	_, err := j.absencesCollection.UpdateByID(ctx, absences.Id, bson.M{"$set": absences})
	return err
}

func (j *journalRepository) GetAbsenceByID(ctx context.Context, id primitive.ObjectID) (absence entities.Absences, err error) {
	err = j.absencesCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&absence)
	return
}

func (j *journalRepository) DeleteAbsenceByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := j.absencesCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
