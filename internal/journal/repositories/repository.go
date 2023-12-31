package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	general "studyum/internal/general/entities"
	"studyum/internal/journal/entities"
	"studyum/pkg/hMongo"
	"studyum/pkg/slicetools"
	"time"
)

type Repository interface {
	GetTypeID(ctx context.Context, studyPlaceID primitive.ObjectID, type_, typeName string) (primitive.ObjectID, error)

	GetMarkByID(ctx context.Context, id primitive.ObjectID) (entities.Mark, error)
	AddMarks(ctx context.Context, marks []entities.StudentMark) error
	AddMark(ctx context.Context, mark entities.StudentMark) error
	UpdateMark(ctx context.Context, mark entities.StudentMark) error
	DeleteMarkByID(ctx context.Context, id primitive.ObjectID) error

	GetAllAvailableOptions(ctx context.Context, id primitive.ObjectID, editable bool) ([]entities.AvailableOption, error)
	GetAvailableOptions(ctx context.Context, id, teacherID primitive.ObjectID, editable bool) ([]entities.AvailableOption, error)
	GetAvailableTuitionOptions(ctx context.Context, id, groupID primitive.ObjectID, editable bool) ([]entities.AvailableOption, error)

	GetStudentJournal(ctx context.Context, userId, groupID primitive.ObjectID, studyPlaceID primitive.ObjectID) (entities.Journal, error)
	GetJournal(ctx context.Context, studyPlaceID primitive.ObjectID, groupID primitive.ObjectID, subjectID primitive.ObjectID) (entities.Journal, error)

	GetLessonByID(ctx context.Context, id primitive.ObjectID) (entities.JournalLesson, error)
	GetStudentLessonByID(ctx context.Context, studentID, id primitive.ObjectID) (entities.JournalLesson, error)

	GetLessons(ctx context.Context, userId primitive.ObjectID, group, teacher, subject string, studyPlaceID primitive.ObjectID) ([]entities.Lesson, error)

	GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID) (general.StudyPlace, error)

	GetAbsenceByID(ctx context.Context, id primitive.ObjectID) (entities.Absence, error)
	AddAbsences(ctx context.Context, absences []entities.Absence) error
	AddAbsence(ctx context.Context, absence entities.Absence) error
	UpdateAbsence(ctx context.Context, absence entities.Absence) error
	DeleteAbsenceByID(ctx context.Context, id primitive.ObjectID) error

	GenerateMarksReport(ctx context.Context, group string, lessonType string, mark string, from, to *time.Time, studyPlaceID primitive.ObjectID) (entities.GeneratedTable, error)
	GenerateAbsencesReport(ctx context.Context, group string, from, to *time.Time, id primitive.ObjectID) (entities.GeneratedTable, error)
}

type repository struct {
	users           *mongo.Collection
	lessons         *mongo.Collection
	studyPlaces     *mongo.Collection
	studyPlaceUsers *mongo.Collection
	marks           *mongo.Collection
	absences        *mongo.Collection

	database *mongo.Database
}

func NewJournalRepository(users *mongo.Collection, lessons *mongo.Collection, studyPlaces *mongo.Collection, studyPlaceUsers *mongo.Collection, marks *mongo.Collection, absences *mongo.Collection, database *mongo.Database) Repository {
	return &repository{users: users, lessons: lessons, studyPlaces: studyPlaces, studyPlaceUsers: studyPlaceUsers, marks: marks, absences: absences, database: database}
}

func (j *repository) GetTypeID(ctx context.Context, studyPlaceID primitive.ObjectID, type_, typeName string) (primitive.ObjectID, error) {
	var collection string
	switch type_ {
	case "teacher":
		var value struct {
			ID primitive.ObjectID `bson:"_id"`
		}

		err := j.database.Collection("Users").FindOne(ctx, bson.M{"studyPlaceInfo.roleName": typeName}).Decode(&value)
		if err != nil {
			return primitive.ObjectID{}, err
		}

		return value.ID, nil
	case "group", "student":
		collection = "Groups"
	case "subject":
		collection = "Subjects"
	case "room":
		collection = "Rooms"
	}

	var value struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	err := j.database.Collection(collection).FindOne(ctx, bson.M{"studyPlaceID": studyPlaceID, type_: typeName}).Decode(&value)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return value.ID, nil
}

func (j *repository) GenerateMarksReport(ctx context.Context, group string, lessonType string, mark string, from, to *time.Time, studyPlaceID primitive.ObjectID) (entities.GeneratedTable, error) {
	var lessonMatcher = bson.M{
		"group":        group,
		"studyPlaceID": studyPlaceID,
		"type":         lessonType,
	}

	if from != nil {
		lessonMatcher["startDate"] = bson.M{"$gte": from}
	}

	if to != nil {
		lessonMatcher["startDate"] = bson.M{"$lte": to}
	}

	if from != nil && to != nil {
		lessonMatcher["startDate"] = bson.M{"$gte": from, "$lte": to}
	}

	var cursor, err = j.users.Aggregate(ctx, bson.A{
		bson.M{
			"$group": bson.M{"_id": nil, "user": bson.M{"$push": "$$ROOT"}},
		},
		bson.M{
			"$project": bson.M{
				"user._id":          1,
				"user.name":         1,
				"user.role":         1,
				"user.roleName":     1,
				"user.studyPlaceID": 1,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "SignUpCodes",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":         1,
							"role":         1,
							"roleName":     1,
							"studyPlaceID": 1,
						},
					},
				},
				"as": "codeUsers",
			},
		},
		bson.M{
			"$project": bson.M{
				"user": bson.M{
					"$filter": bson.M{
						"input": bson.M{"$concatArrays": bson.A{"$codeUsers", "$user"}},
						"as":    "user",
						"cond":  bson.M{"$and": bson.A{bson.M{"$eq": bson.A{"$$user.role", "group"}}, bson.M{"$eq": bson.A{"$$user.roleName", group}}, bson.M{"$eq": bson.A{"$$user.studyPlaceID", studyPlaceID}}}},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "Lessons",
				"let":  bson.M{"userID": "$_id"},
				"pipeline": bson.A{
					bson.M{
						"$match": lessonMatcher,
					},
				},
				"as": "lessons",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "StudyPlaces",
				"pipeline": bson.A{
					bson.M{"$match": bson.M{
						"_id": studyPlaceID,
					}},
				},
				"as": "studyPlace",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"lessons": bson.M{
					"$sortArray": bson.M{
						"input":  "$lessons",
						"sortBy": bson.M{"startDate": 1},
					},
				},
			},
		},
		bson.M{
			"$unwind": "$user",
		},
		bson.M{
			"$unwind": "$lessons",
		},
		bson.M{
			"$addFields": bson.M{
				"lessons.marks": hMongo.Filter("lessons.marks", hMongo.AEq("$$marks.studentID", "$user._id")),
			},
		},
		bson.M{
			"$addFields": bson.M{
				"lessons.marks": hMongo.Filter("lessons.marks", hMongo.AEq("$$marks.mark", mark)),
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"user":    "$user",
					"subject": "$lessons.subject",
				},
				"title":   bson.M{"$first": "$user.name"},
				"subject": bson.M{"$first": "$lessons.subject"},
				"marks":   bson.M{"$push": "$lessons.marks"},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":  nil,
				"list": bson.M{"$push": "$$ROOT"},
			},
		},
		bson.M{
			"$replaceRoot": bson.D{
				{"newRoot", bson.D{hMongo.Func(`function (list) {
					const groupBy = function (xs, key) {
						return xs.reduce(function (rv, x) {
							(rv[x[key]] = rv[x[key]] || []).push(x);
							return rv;
						}, {});
					};
                    
    				list.forEach(el => {
                        el.totalLen = el.marks.length
                        el.marksLen = el.marks.flatMap(l => l ?? []).length
                                               
                        el.text = (el.totalLen - el.marksLen).toString() + "/" + el.totalLen.toString()
                        
                        delete el._id
                        delete el.marks
                    })
                    
                    list = groupBy(list, "title")
                    list = Object.entries(list).map(entry => entry[1].sort((a, b) => a.subject > b.subject))
                    list.forEach(el => el[0].temp = el.reduce((r, e) => {
                        r.totalLessonsLen += e.totalLen
                        r.totalMarksLen += e.marksLen
                        return r
                    }, {totalLessonsLen: 0, totalMarksLen: 0}))

                  	let titles = list[0].map(el => el.subject)
                  	titles.unshift("")
                  	titles.push("")
                  	
                  	list = list.map(el => {
                          let temp = el[0].temp
                          let text = (temp.totalLessonsLen - temp.totalMarksLen).toString() + "/" + temp.totalLessonsLen.toString()
                          return [el[0].title, ...el.map(v => v.text), text]
                    })
                                                           
   					return {titles: titles, rows: list}
				}`, "$list")}},
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":  0,
				"list": 0,
			},
		},
	})

	if err != nil {
		return entities.GeneratedTable{}, err
	}

	if !cursor.Next(ctx) {
		return entities.GeneratedTable{}, nil
	}
	var table entities.GeneratedTable
	if err = cursor.Decode(&table); err != nil {
		return entities.GeneratedTable{}, err
	}

	return table, nil
}

func (j *repository) GenerateAbsencesReport(ctx context.Context, group string, from, to *time.Time, studyPlaceID primitive.ObjectID) (entities.GeneratedTable, error) {
	var lessonMatcher = bson.M{
		"group":        group,
		"studyPlaceID": studyPlaceID,
	}

	if from != nil {
		lessonMatcher["startDate"] = bson.M{"$gte": from}
	}

	if to != nil {
		lessonMatcher["startDate"] = bson.M{"$lte": to}
	}

	if from != nil && to != nil {
		lessonMatcher["startDate"] = bson.M{"$gte": from, "$lte": to}
	}

	var cursor, err = j.users.Aggregate(ctx, bson.A{
		bson.M{
			"$group": bson.M{"_id": nil, "user": bson.M{"$push": "$$ROOT"}},
		},
		bson.M{
			"$project": bson.M{
				"user._id":          1,
				"user.name":         1,
				"user.role":         1,
				"user.roleName":     1,
				"user.studyPlaceID": 1,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "SignUpCodes",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":         1,
							"role":         1,
							"roleName":     1,
							"studyPlaceID": 1,
						},
					},
				},
				"as": "codeUsers",
			},
		},
		bson.M{
			"$project": bson.M{
				"user": bson.M{
					"$filter": bson.M{
						"input": bson.M{"$concatArrays": bson.A{"$codeUsers", "$user"}},
						"as":    "user",
						"cond":  bson.M{"$and": bson.A{bson.M{"$eq": bson.A{"$$user.role", "group"}}, bson.M{"$eq": bson.A{"$$user.roleName", group}}, bson.M{"$eq": bson.A{"$$user.studyPlaceID", studyPlaceID}}}},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "Lessons",
				"let":  bson.M{"userID": "$_id"},
				"pipeline": bson.A{
					bson.M{"$match": lessonMatcher},
				},
				"as": "lessons",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "StudyPlaces",
				"pipeline": bson.A{
					bson.M{"$match": bson.M{
						"_id": studyPlaceID,
					}},
				},
				"as": "studyPlace",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"lessons": bson.M{
					"$sortArray": bson.M{
						"input":  "$lessons",
						"sortBy": bson.M{"startDate": 1},
					},
				},
			},
		},
		bson.M{
			"$unwind": "$user",
		},
		bson.M{
			"$unwind": "$lessons",
		},
		bson.M{
			"$addFields": bson.M{
				"lessons.absences": hMongo.Filter("lessons.absences", hMongo.AEq("$$absences.studentID", "$user._id")),
			},
		},
		bson.M{
			"$addFields": bson.M{
				"lessons.absences": hMongo.Filter("lessons.absences", hMongo.AEq("$$absences.time", nil)),
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"user": "$user",
					"date": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$lessons.startDate"}},
				},
				"title":    bson.M{"$first": "$user.name"},
				"date":     bson.M{"$first": "$lessons.startDate"},
				"absences": bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$isArray": "$lessons.absences"}, "then": bson.M{"$size": "$lessons.absences"}, "else": 0}}}},
		},
		bson.M{
			"$group": bson.M{
				"_id":  nil,
				"list": bson.M{"$push": "$$ROOT"},
			},
		},
		bson.M{
			"$replaceRoot": bson.D{
				{"newRoot", bson.D{hMongo.Func(`function (list) {
					const groupBy = function (xs, key) {
						return xs.reduce(function (rv, x) {
							(rv[x[key]] = rv[x[key]] || []).push(x);
							return rv;
						}, {});
					};
                    
                    list = groupBy(list, "title")
                    list = Object.entries(list).map(entry => entry[1].sort((a, b) => a.date > b.date))

                    let titles = list[0].map(el => el.date.getDate() + "." + (el.date.getMonth() + 1))
                  	titles.unshift("")
                  	titles.push("")
                    
                    list = list.map(el => {
                          return [el[0].title, ...el.map(v => v.absences === 0 ? "" : v.absences.toString()), ""]
                    })
                    
					return {titles: titles, rows: list}
				}`, "$list")}},
			},
		},
	})

	if err != nil {
		return entities.GeneratedTable{}, err
	}

	if !cursor.Next(ctx) {
		return entities.GeneratedTable{}, nil
	}
	var table entities.GeneratedTable
	if err = cursor.Decode(&table); err != nil {
		return entities.GeneratedTable{}, err
	}

	return table, nil
}

func (j *repository) getAvailableOptions(ctx context.Context, matcher bson.M, editable bool) ([]entities.AvailableOption, error) {
	aggregate, err := j.lessons.Aggregate(ctx, bson.A{
		bson.M{"$match": matcher},
		bson.M{"$group": bson.M{
			"_id": bson.M{
				"teacherID": "$teacherID",
				"subjectID": "$subjectID",
				"groupID":   "$groupID",
			},
			"teacherID": bson.M{"$first": "$teacherID"},
			"subjectID": bson.M{"$first": "$subjectID"},
			"groupID":   bson.M{"$first": "$groupID"},
		}},
		bson.M{
			"$lookup": bson.M{
				"from":         "Teachers",
				"localField":   "teacherID",
				"foreignField": "_id",
				"as":           "teacher",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "Groups",
				"localField":   "groupID",
				"foreignField": "_id",
				"as":           "group",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "Subjects",
				"localField":   "subjectID",
				"foreignField": "_id",
				"as":           "subject",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"teacher": bson.M{"$first": "$teacher"},
				"group":   bson.M{"$first": "$group"},
				"subject": bson.M{"$first": "$subject"},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "StudyPlaceUsers",
				"let":  bson.M{"groupID": "$groupID"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$and": bson.A{
							bson.M{"$eq": bson.A{"$studyPlaceID", matcher["studyPlaceID"]}},
							bson.M{"$eq": bson.A{"$$groupID", "$roleID"}},
						}}},
					},
				},
				"as": "hasMembers",
			},
		},
		bson.M{"$addFields": bson.M{
			"editable":   editable,
			"hasMembers": bson.M{"$ne": bson.A{bson.M{"$size": "$hasMembers"}, 0}},
		}},
		bson.M{"$sort": bson.M{"hasMembers": -1, "group": 1, "groupID": 1, "subject": 1, "subjectID": 1, "teacher": 1, "teacherID": 1}},
	})
	if err != nil {
		return nil, err
	}

	var options []entities.AvailableOption
	if err = aggregate.All(ctx, &options); err != nil {
		return nil, err
	}

	return options, nil
}

func (j *repository) GetAllAvailableOptions(ctx context.Context, id primitive.ObjectID, editable bool) ([]entities.AvailableOption, error) {
	return j.getAvailableOptions(ctx, bson.M{"studyPlaceID": id}, editable)
}

func (j *repository) GetAvailableOptions(ctx context.Context, id, teacherID primitive.ObjectID, editable bool) ([]entities.AvailableOption, error) {
	return j.getAvailableOptions(ctx, bson.M{"studyPlaceID": id, "teacherID": teacherID}, editable)
}

func (j *repository) GetAvailableTuitionOptions(ctx context.Context, id, groupID primitive.ObjectID, editable bool) ([]entities.AvailableOption, error) {
	return j.getAvailableOptions(ctx, bson.M{"studyPlaceID": id, "groupID": groupID}, editable)
}

func (j *repository) GetStudentJournal(ctx context.Context, userId, groupID primitive.ObjectID, studyPlaceID primitive.ObjectID) (entities.Journal, error) {
	cursor, err := j.lessons.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"studyPlaceID": studyPlaceID,
				"groupID":      groupID,
			},
		},
		{
			"$lookup": bson.M{
				"from": "AggregatedMarks",
				"let": bson.M{
					"lessonID": "$_id",
				},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$$lessonID", "$lessonID"}},
									bson.M{"$eq": bson.A{"$studentID", userId}},
								},
							},
						},
					},
					bson.M{
						"$project": bson.M{
							"_id":        1,
							"mark":       1,
							"markWeight": 1,
						},
					},
				},
				"as": "marks",
			},
		},
		{
			"$lookup": bson.M{
				"from": "Absences",
				"let": bson.M{
					"lessonID": "$_id",
				},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$lessonID", "$$lessonID"}},
									bson.M{"$eq": bson.A{"$studentID", userId}},
								},
							},
						},
					},
					bson.M{
						"$project": bson.M{
							"_id":  1,
							"time": 1,
						},
					},
				},
				"as": "absences",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "Subjects",
				"localField":   "subjectID",
				"foreignField": "_id",
				"as":           "subjects",
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"subject": bson.M{
						"_id":   bson.M{"$first": "$subjects._id"},
						"title": bson.M{"$first": "$subjects.subject"},
					},
					"date": bson.M{
						"date": bson.M{
							"$dateTrunc": bson.M{
								"date": "$startDate",
								"unit": "day",
							},
						},
					},
				},
				"entries": bson.M{"$push": bson.M{
					"lessonID": "$_id",
					"typeID":   "$typeID",
					"marks":    "$marks",
					"absences": "$absences",
				}},
			},
		},
		{
			"$project": bson.M{
				"cell": bson.M{
					"_id":      "$_id",
					"date":     "$_id.date",
					"rowTitle": "$_id.subject",
					"entries":  "$entries",
				},
				"cell.date.typeIDs": "$entries.typeIDs",
			},
		},
		{
			"$group": bson.M{
				"_id":       nil,
				"dates":     bson.M{"$addToSet": "$_id.date"},
				"rowTitles": bson.M{"$addToSet": "$_id.subject"},
				"cells":     bson.M{"$push": "$cell"},
			},
		},
		{
			"$lookup": bson.M{
				"from": "JournalConfigs",
				"pipeline": []bson.M{
					{
						"$match": bson.M{"studyPlaceID": studyPlaceID},
					},
				},
				"as": "info.configs",
			},
		},
		{
			"$project": bson.M{
				"_id":       0,
				"dates":     1,
				"rowTitles": 1,
				"cells":     1,
				"info":      1,
			},
		},
	})
	if err != nil {
		return entities.Journal{}, err
	}

	if !cursor.Next(ctx) {
		return entities.Journal{}, nil
	}
	var journal entities.Journal
	if err = cursor.Decode(&journal); err != nil {
		return entities.Journal{}, err
	}

	return journal, nil
}

func (j *repository) GetJournal(ctx context.Context, studyPlaceID primitive.ObjectID, groupID primitive.ObjectID, subjectID primitive.ObjectID) (entities.Journal, error) {
	var cursor, err = j.studyPlaceUsers.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"studyPlaceID": studyPlaceID,
				"role":         "student",
				"roleID":       groupID,
			},
		},
		{
			"$lookup": bson.M{
				"from": "Lessons",
				"let":  bson.M{"groupID": groupID, "subjectID": subjectID},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{studyPlaceID, "$studyPlaceID"}},
									bson.M{"$eq": bson.A{"$groupID", "$$groupID"}},
									bson.M{"$eq": bson.A{"$subjectID", "$$subjectID"}},
								},
							},
						},
					},
				},
				"as": "lessons",
			},
		},
		{
			"$project": bson.M{
				"_id":     1,
				"lessons": 1,
				"name":    1,
			},
		},
		{
			"$unwind": "$lessons",
		},
		{
			"$lookup": bson.M{
				"from": "AggregatedMarks",
				"let":  bson.M{"lessonID": "$lessons._id", "studentID": "$_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$$lessonID", "$lessonID"}},
									bson.M{"$eq": bson.A{"$$studentID", "$studentID"}},
								},
							},
						},
					},
					{
						"$project": bson.M{
							"_id":        1,
							"markID":     1,
							"mark":       1,
							"markWeight": 1,
						},
					},
				},
				"as": "marks",
			},
		},
		{
			"$lookup": bson.M{
				"from": "Absences",
				"let":  bson.M{"lessonID": "$lessons._id", "studentID": "$_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$$lessonID", "$lessonID"}},
									bson.M{"$eq": bson.A{"$$studentID", "$studentID"}},
								},
							},
						},
					},
					{
						"$project": bson.M{
							"_id":  1,
							"time": 1,
						},
					},
				},
				"as": "absences",
			},
		},
		{
			"$project": bson.M{
				"date": bson.M{
					"_id":     "$lessons._id",
					"date":    "$lessons.startDate",
					"typeIDs": bson.A{"$lessons.typeID"},
				},
				"rowTitle": bson.M{
					"_id":   "$_id",
					"title": "$name",
				},
				"cell": bson.M{
					"date": bson.M{
						"_id":  "$lessons._id",
						"date": "$lessons.startDate",
					},
					"rowTitle": bson.M{
						"_id":   "$_id",
						"title": "$name",
					},
					"entries": bson.A{
						bson.M{
							"lessonID": "$lessons._id",
							"typeID":   "$lessons.typeID",
							"marks":    "$marks",
							"absences": "$absences",
						},
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id":       nil,
				"dates":     bson.M{"$addToSet": "$date"},
				"rowTitles": bson.M{"$addToSet": "$rowTitle"},
				"cells":     bson.M{"$push": "$cell"},
			},
		},
		{
			"$lookup": bson.M{
				"from": "JournalConfigs",
				"pipeline": []bson.M{
					{
						"$match": bson.M{"studyPlaceID": studyPlaceID},
					},
				},
				"as": "info.configs",
			},
		},
	})

	if err != nil {
		return entities.Journal{}, err
	}

	if !cursor.Next(ctx) {
		return entities.Journal{}, nil
	}

	var journal entities.Journal
	if err = cursor.Decode(&journal); err != nil {
		return entities.Journal{}, err
	}

	return journal, nil
}

func (j *repository) GetLessonByID(ctx context.Context, id primitive.ObjectID) (lesson entities.JournalLesson, err error) {
	cursor, err := j.lessons.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"_id": id}},
		bson.M{"$lookup": bson.M{"from": "StudyPlaceUsers", "localField": "teacherID", "foreignField": "_id", "as": "teacher"}},
		bson.M{"$lookup": bson.M{"from": "Groups", "localField": "groupID", "foreignField": "_id", "as": "group"}},
		bson.M{"$lookup": bson.M{"from": "Subjects", "localField": "subjectID", "foreignField": "_id", "as": "subject"}},
		bson.M{"$lookup": bson.M{"from": "Rooms", "localField": "roomID", "foreignField": "_id", "as": "room"}},
		bson.M{"$lookup": bson.M{"from": "LessonTypes", "localField": "typeID", "foreignField": "_id", "as": "type"}},
		bson.M{
			"$addFields": bson.M{
				"subject": bson.M{"$first": "$subject.subject"},
				"room":    bson.M{"$first": "$room.room"},
				"teacher": bson.M{"$first": "$teacher.roleName"},
				"group":   bson.M{"$first": "$group.group"},
				"type":    bson.M{"$first": "$type"},
			},
		},
	})
	if err != nil {
		return entities.JournalLesson{}, err
	}

	if !cursor.Next(ctx) {
		return entities.JournalLesson{}, mongo.ErrNoDocuments
	}

	if err = cursor.Decode(&lesson); err != nil {
		return entities.JournalLesson{}, err
	}

	return
}

func (j *repository) GetStudentLessonByID(ctx context.Context, studentID, id primitive.ObjectID) (lesson entities.JournalLesson, err error) {
	cursor, err := j.lessons.Aggregate(ctx, []bson.M{
		{"$match": bson.M{"_id": id}},
		{"$lookup": bson.M{"from": "Teachers", "localField": "teacherID", "foreignField": "_id", "as": "teacher"}},
		{"$lookup": bson.M{"from": "Groups", "localField": "groupID", "foreignField": "_id", "as": "group"}},
		{"$lookup": bson.M{"from": "Subjects", "localField": "subjectID", "foreignField": "_id", "as": "subject"}},
		{"$lookup": bson.M{"from": "Rooms", "localField": "roomID", "foreignField": "_id", "as": "room"}},
		{"$lookup": bson.M{"from": "LessonTypes", "localField": "typeID", "foreignField": "_id", "as": "type"}},
		{
			"$lookup": bson.M{
				"from": "AggregatedMarks",
				"let":  bson.M{"lessonID": "$_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$$lessonID", "$lessonID"}},
									bson.M{"$eq": bson.A{"$studentID", studentID}},
								},
							},
						},
					},
					{
						"$project": bson.M{
							"_id":        1,
							"markID":     1,
							"mark":       1,
							"markWeight": 1,
						},
					},
				},
				"as": "marks",
			},
		},
		{
			"$lookup": bson.M{
				"from": "Absences",
				"let":  bson.M{"lessonID": "$_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$$lessonID", "$lessonID"}},
									bson.M{"$eq": bson.A{"$studentID", studentID}},
								},
							},
						},
					},
					{
						"$project": bson.M{
							"_id":  1,
							"time": 1,
						},
					},
				},
				"as": "absence",
			},
		},
		{
			"$addFields": bson.M{
				"subject": bson.M{"$first": "$subject"},
				"room":    bson.M{"$first": "$room"},
				"teacher": bson.M{"$first": "$teacher"},
				"group":   bson.M{"$first": "$group"},
				"type":    bson.M{"$first": "$type"},
				"absence": bson.M{"$first": "$absence"},
			},
		},
		{
			"$lookup": bson.M{
				"from": "StudyPlaceMarks",
				"let":  bson.M{"typeID": "$type._id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{"$expr": bson.M{
							"$in": bson.A{"$$typeID", "$assignLessonTypeIDs"},
						}},
					},
				},
				"as": "type.availableMarks",
			},
		},
	})
	if err != nil {
		return entities.JournalLesson{}, err
	}

	if !cursor.Next(ctx) {
		return entities.JournalLesson{}, mongo.ErrNoDocuments
	}

	if err = cursor.Decode(&lesson); err != nil {
		return entities.JournalLesson{}, err
	}

	return
}

func (j *repository) GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID) (studyPlace general.StudyPlace, err error) {
	err = j.studyPlaces.FindOne(ctx, bson.M{"_id": id}).Decode(&studyPlace)
	return
}

func (j *repository) GetLessons(ctx context.Context, userId primitive.ObjectID, group, teacher, subject string, studyPlaceID primitive.ObjectID) ([]entities.Lesson, error) {
	lessonsCursor, err := j.lessons.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$lookup", bson.M{
			"from":         "Marks",
			"localField":   "_id",
			"foreignField": "lessonId",
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"studentID": userId}}},
			},
			"as": "marks",
		}}},
		bson.D{{"$match", bson.M{"group": group, "teacher": teacher, "subject": subject, "studyPlaceID": studyPlaceID}}},
		bson.D{{"$sort", bson.M{"date": 1}}},
	})
	if err != nil {
		return nil, err
	}

	var marks []entities.Lesson
	if err = lessonsCursor.All(ctx, &marks); err != nil {
		return nil, err
	}

	return marks, nil
}

func (j *repository) AddMarks(ctx context.Context, marks []entities.StudentMark) error {
	_, err := j.marks.InsertMany(ctx, slicetools.ToInterface(marks))
	return err
}

func (j *repository) AddMark(ctx context.Context, mark entities.StudentMark) error {
	_, err := j.marks.InsertOne(ctx, mark)
	return err
}

func (j *repository) UpdateMark(ctx context.Context, mark entities.StudentMark) error {
	_, err := j.marks.UpdateOne(ctx, bson.M{"_id": mark.ID}, mark)
	return err
}

func (j *repository) DeleteMarkByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := j.marks.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (j *repository) GetMarkByID(ctx context.Context, id primitive.ObjectID) (mark entities.Mark, err error) {
	err = j.marks.FindOne(ctx, bson.M{"_id": id}).Decode(&mark)
	return
}

func (j *repository) GetAbsenceByID(ctx context.Context, id primitive.ObjectID) (absence entities.Absence, err error) {
	err = j.absences.FindOne(ctx, bson.M{"_id": id}).Decode(&absence)
	return
}

func (j *repository) AddAbsences(ctx context.Context, absences []entities.Absence) error {
	_, err := j.absences.InsertMany(ctx, slicetools.ToInterface(absences))
	return err
}

func (j *repository) AddAbsence(ctx context.Context, absence entities.Absence) error {
	_, err := j.absences.InsertOne(ctx, absence)
	return err
}

func (j *repository) UpdateAbsence(ctx context.Context, absence entities.Absence) error {
	_, err := j.absences.UpdateOne(ctx, bson.M{"_id": absence.ID}, absence)
	return err
}

func (j *repository) DeleteAbsenceByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := j.absences.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
