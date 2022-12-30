package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/entities"
	"studyum/pkg/hMongo"
	"time"
)

type JournalRepository interface {
	AddMarks(ctx context.Context, marks []entities.Mark, teacher string) error
	AddMark(ctx context.Context, mark entities.Mark, teacher string) (primitive.ObjectID, error)
	UpdateMark(ctx context.Context, mark entities.Mark, teacher string) error
	DeleteMarkByID(ctx context.Context, id primitive.ObjectID, teacher string) error

	GetAvailableOptions(ctx context.Context, teacher string, editable bool) ([]entities.JournalAvailableOption, error)
	GetAvailableTuitionOptions(ctx context.Context, name string, editable bool) ([]entities.JournalAvailableOption, error)

	GetStudentJournal(ctx context.Context, userId primitive.ObjectID, group string, studyPlaceId primitive.ObjectID) (entities.Journal, error)
	GetJournal(ctx context.Context, option entities.JournalAvailableOption, studyPlaceId primitive.ObjectID) (entities.Journal, error)

	GetLessonByID(ctx context.Context, id primitive.ObjectID) (entities.Lesson, error)
	GetLessons(ctx context.Context, userId primitive.ObjectID, group, teacher, subject string, studyPlaceId primitive.ObjectID) ([]entities.Lesson, error)

	AddAbsences(ctx context.Context, absences []entities.Absence, teacher string) error
	AddAbsence(ctx context.Context, absence entities.Absence, teacher string) (primitive.ObjectID, error)
	UpdateAbsence(ctx context.Context, absence entities.Absence, teacher string) error
	DeleteAbsenceByID(ctx context.Context, id primitive.ObjectID, teacher string) error

	Generate(ctx context.Context, group string, lessonType string, mark string, from, to *time.Time, studyPlaceId primitive.ObjectID) (entities.GeneratedTable, error)
	GenerateAbsences(ctx context.Context, group string, from, to *time.Time, id primitive.ObjectID) (entities.GeneratedTable, error)
}

type journalRepository struct {
	*Repository
}

func NewJournalRepository(repository *Repository) JournalRepository {
	return &journalRepository{Repository: repository}
}

func (j *journalRepository) Generate(ctx context.Context, group string, lessonType string, mark string, from, to *time.Time, studyPlaceId primitive.ObjectID) (entities.GeneratedTable, error) {
	var lessonMatcher = bson.M{
		"group":        group,
		"studyPlaceId": studyPlaceId,
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

	var cursor, err = j.usersCollection.Aggregate(ctx, bson.A{
		bson.M{
			"$group": bson.M{"_id": nil, "users": bson.M{"$push": "$$ROOT"}},
		},
		bson.M{
			"$project": bson.M{
				"users._id":          1,
				"users.name":         1,
				"users.type":         1,
				"users.typename":     1,
				"users.studyPlaceID": 1,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "SignUpCodes",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":         1,
							"type":         1,
							"typename":     1,
							"studyPlaceID": 1,
						},
					},
				},
				"as": "codeUsers",
			},
		},
		bson.M{
			"$project": bson.M{
				"users": bson.M{
					"$filter": bson.M{
						"input": bson.M{"$concatArrays": bson.A{"$codeUsers", "$users"}},
						"as":    "user",
						"cond":  bson.M{"$and": bson.A{bson.M{"$eq": bson.A{"$$user.type", "group"}}, bson.M{"$eq": bson.A{"$$user.typename", group}}, bson.M{"$eq": bson.A{"$$user.studyPlaceID", studyPlaceId}}}},
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
						"_id": studyPlaceId,
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
			"$unwind": "$users",
		},
		bson.M{
			"$unwind": "$lessons",
		},
		bson.M{
			"$addFields": bson.M{
				"lessons.marks": hMongo.Filter("lessons.marks", hMongo.AEq("$$marks.studentID", "$users._id")),
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
					"user":    "$users",
					"subject": "$lessons.subject",
				},
				"title":   bson.M{"$first": "$users.name"},
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

func (j *journalRepository) GenerateAbsences(ctx context.Context, group string, from, to *time.Time, studyPlaceId primitive.ObjectID) (entities.GeneratedTable, error) {
	var lessonMatcher = bson.M{
		"group":        group,
		"studyPlaceId": studyPlaceId,
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

	var cursor, err = j.usersCollection.Aggregate(ctx, bson.A{
		bson.M{
			"$group": bson.M{"_id": nil, "users": bson.M{"$push": "$$ROOT"}},
		},
		bson.M{
			"$project": bson.M{
				"users._id":          1,
				"users.name":         1,
				"users.type":         1,
				"users.typename":     1,
				"users.studyPlaceID": 1,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "SignUpCodes",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":         1,
							"type":         1,
							"typename":     1,
							"studyPlaceID": 1,
						},
					},
				},
				"as": "codeUsers",
			},
		},
		bson.M{
			"$project": bson.M{
				"users": bson.M{
					"$filter": bson.M{
						"input": bson.M{"$concatArrays": bson.A{"$codeUsers", "$users"}},
						"as":    "user",
						"cond":  bson.M{"$and": bson.A{bson.M{"$eq": bson.A{"$$user.type", "group"}}, bson.M{"$eq": bson.A{"$$user.typename", group}}, bson.M{"$eq": bson.A{"$$user.studyPlaceID", studyPlaceId}}}},
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
						"_id": studyPlaceId,
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
			"$unwind": "$users",
		},
		bson.M{
			"$unwind": "$lessons",
		},
		bson.M{
			"$addFields": bson.M{
				"lessons.absences": hMongo.Filter("lessons.absences", hMongo.AEq("$$absences.studentID", "$users._id")),
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
					"user": "$users",
					"date": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$lessons.startDate"}},
				},
				"title":    bson.M{"$first": "$users.name"},
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
		bson.M{"$lookup": bson.M{
			"from": "Users",
			"let":  bson.M{"group": "$group"},
			"pipeline": bson.A{
				bson.M{
					"$group": bson.M{"_id": nil, "users": bson.M{"$push": "$$ROOT"}},
				},
				bson.M{
					"$lookup": bson.M{
						"from": "SignUpCodes",
						"pipeline": bson.A{
							bson.M{
								"$project": bson.M{
									"name":         1,
									"type":         1,
									"typename":     1,
									"studyPlaceID": 1,
								},
							},
						},
						"as": "codeUsers",
					},
				},
				bson.M{
					"$project": bson.M{
						"len": bson.M{"$size": bson.M{
							"$filter": bson.M{
								"input": bson.M{"$concatArrays": bson.A{"$codeUsers", "$users"}},
								"as":    "user",
								"cond": bson.M{
									"$and": bson.A{
										bson.M{"$eq": bson.A{"$$user.typename", "$$group"}},
									},
								},
							},
						}},
					},
				},
			},
			"as": "users",
		}},
		bson.M{"$addFields": bson.M{"temp": bson.M{"$first": "$users"}}},
		bson.M{"$match": bson.M{"temp.len": bson.M{"$gt": 0}}},
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

func (j *journalRepository) GetAvailableTuitionOptions(ctx context.Context, group string, editable bool) ([]entities.JournalAvailableOption, error) {
	aggregate, err := j.lessonsCollection.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"group": group}},
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

func (j *journalRepository) getCellColor() string {
	return `function (studyPlace, lesson) {
                        if (!lesson) return "";

                        let color = studyPlace.journalColors.general
						let type = studyPlace.lessonTypes.find(v => v.type === lesson.type);
						if (type === undefined) return studyPlace.journalColors.general;

                        for (let mark of lesson.marks ?? []) {
                            let markType = type.marks.find(m => m.mark === mark.mark);
                            if (markType === undefined || markType.workOutTime === undefined || markType.workOutTime === 0) return studyPlace.journalColors.general;

                            let date = new Date(lesson.startDate);
							date.setSeconds(lesson.startDate.getSeconds() + markType.workOutTime);
                            color = date.getTime() > new Date().getTime() ? studyPlace.journalColors.warning : studyPlace.journalColors.danger;
                        }

						if (!lesson.absences || lesson.absences.length === 0 || type.absenceWorkOutTime === undefined || type.absenceWorkOutTime === 0) return color

						let date = new Date(lesson.startDate);
						date.setSeconds(lesson.startDate.getSeconds() + type.absenceWorkOutTime);
                        return date.getTime() > new Date().getTime() ? studyPlace.journalColors.warning : studyPlace.journalColors.danger;
                    }`
}

func (j *journalRepository) getRowInfo() string {
	return `function(studyPlace, row) {
		let marks = row.lessons.flatMap(l => l?.marks ?? [])

		let markList = studyPlace.lessonTypes
				.flatMap(t => t.marks.concat(t.standaloneMarks ?? [])).map(m => m.mark)
				.filter((v, i, a) => a.indexOf(v) === i)
						
		let marksAmount = {}
		for (let m of markList) {
		  marksAmount[m] = {
              mark: m,
              amount: 0
		  }
		}
        for (let mark of marks) marksAmount[mark.mark].amount++
        row.marksAmount = Object.values(marksAmount)

        marks = marks.map(m => Number.parseInt(m.mark)).filter(m => m)
        
        row.numericMarksSum = marks.reduce((sum, a) => sum + a, 0)
		row.numericMarksAmount = marks.length

		let absences = row.lessons.flatMap(l => l?.absences ?? []).map(a => a.time)
		row.absencesAmount = absences.filter(v => !v).length
        row.absencesTime = absences.filter(v => v).reduce((sum, a) => sum + a, 0)
		
		let color = studyPlace.journalColors.general
		for (let lesson of row.lessons) {
		  if (lesson == null) continue
		
		  if (lesson.journalCellColor == studyPlace.journalColors.warning)
			  color = studyPlace.journalColors.warning
		
		  if (lesson.journalCellColor == studyPlace.journalColors.danger){
			  color = studyPlace.journalColors.danger
			  break
		  }
		}

		row.color = color
		return row
	}`
}

func (j *journalRepository) GetStudentJournal(ctx context.Context, userId primitive.ObjectID, group string, studyPlaceId primitive.ObjectID) (entities.Journal, error) {
	cursor, err := j.lessonsCollection.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"group": group, "studyPlaceId": studyPlaceId}},
		bson.M{"$addFields": bson.M{
			"marks":    hMongo.Filter("marks", hMongo.AEq("$$marks.studentID", userId)),
			"absences": hMongo.Filter("absences", hMongo.AEq("$$absences.studentID", userId)),
		}},
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
						"body": j.getCellColor(),
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
                            rows.unshift({title: key, lessons: [], info: {}})

                            let added = 0
                            for (let i = 0; i < value.length; i++) {
                                let startTime = new Date(value[i].startDate.toDateString()).getTime()
                                if (i > 0 && new Date(value[i - 1].startDate.toDateString()).getTime() === startTime) {
                                    let prevLesson = rows[0].lessons.at(-1)

                                    if (value[i].journalCellColor != studyPlace.journalColors.general && prevLesson.journalCellColor == studyPlace.journalColors.general) {
                                        prevLesson.journalCellColor = value[i].journalCellColor
                                    }

                                    prevLesson.marks = prevLesson.marks?.concat(value[i].marks ?? []) ?? value[i].marks
                                    prevLesson.absences = prevLesson.absences?.concat(value[i].absences ?? []) ?? value[i].absences
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

                        return rows.sort((a, b) => a.title > b.title)
                    }`,
						"args": bson.A{"$studyPlace", "$lessons", "$dates"},
						"lang": "js",
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"rows": hMongo.Map("rows", bson.D{
					hMongo.Func(j.getRowInfo(), "$studyPlace", "$$rows"),
				}),
			},
		}, //TODO add to sql
		bson.M{
			"$addFields": bson.M{
				"info": bson.M{
					"editable":   false,
					"studyPlace": "$studyPlace",
					"group":      group,
					"teacher":    "",
					"subject":    "",
					"time":       time.Now(),
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

func (j *journalRepository) GetJournal(ctx context.Context, option entities.JournalAvailableOption, studyPlaceId primitive.ObjectID) (entities.Journal, error) {
	var cursor, err = j.usersCollection.Aggregate(ctx, bson.A{
		bson.M{
			"$group": bson.M{"_id": nil, "users": bson.M{"$push": "$$ROOT"}},
		},
		bson.M{
			"$project": bson.M{
				"users._id":          1,
				"users.name":         1,
				"users.type":         1,
				"users.typename":     1,
				"users.studyPlaceID": 1,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "SignUpCodes",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":         1,
							"type":         1,
							"typename":     1,
							"studyPlaceID": 1,
						},
					},
				},
				"as": "codeUsers",
			},
		},
		bson.M{
			"$project": bson.M{
				"users": bson.M{
					"$filter": bson.M{
						"input": bson.M{"$concatArrays": bson.A{"$codeUsers", "$users"}},
						"as":    "user",
						"cond":  bson.M{"$and": bson.A{bson.M{"$eq": bson.A{"$$user.type", "group"}}, bson.M{"$eq": bson.A{"$$user.typename", option.Group}}, bson.M{"$eq": bson.A{"$$user.studyPlaceID", studyPlaceId}}}},
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
						"$match": bson.M{
							"subject":      option.Subject,
							"group":        option.Group,
							"teacher":      option.Teacher,
							"studyPlaceId": studyPlaceId,
						},
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
						"_id": studyPlaceId,
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
			"$unwind": "$users",
		},
		bson.M{
			"$unwind": "$lessons",
		},
		bson.M{
			"$addFields": bson.M{
				"lessons.marks":    hMongo.Filter("lessons.marks", hMongo.AEq("$$marks.studentID", "$users._id")),
				"lessons.absences": hMongo.Filter("lessons.absences", hMongo.AEq("$$absences.studentID", "$users._id")),
			},
		},
		bson.M{
			"$addFields": bson.M{
				"lessons.journalCellColor": bson.M{
					"$function": bson.M{
						"body": j.getCellColor(),
						"args": bson.A{bson.M{"$first": "$studyPlace"}, "$lessons"},
						"lang": "js",
					},
				},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"_id":   "$users._id",
					"title": "$users.name",
				},
				"lessons":    bson.M{"$push": "$lessons"},
				"studyPlace": bson.M{"$first": bson.M{"$first": "$studyPlace"}},
			},
		},
		bson.M{
			"$project": bson.M{
				"row": bson.M{
					"_id":     "$_id._id",
					"title":   "$_id.title",
					"lessons": "$lessons",
				},
				"studyPlace": "$studyPlace",
			},
		},
		bson.M{
			"$addFields": bson.D{
				{"row", bson.D{hMongo.Func(j.getRowInfo(), "$studyPlace", "$row")}},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":        nil,
				"dates":      bson.M{"$first": "$row.lessons"},
				"rows":       bson.M{"$push": "$row"},
				"studyPlace": bson.M{"$first": "$studyPlace"},
			},
		},
		bson.M{
			"$project": bson.M{
				"dates.marks":            0,
				"dates.journalCellColor": 0,
				"dates.studyPlace":       0,
				"rows.studyPlace":        0,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"info": bson.M{
					"editable":   option.Editable,
					"studyPlace": "$studyPlace",
					"subject":    option.Subject,
					"group":      option.Group,
					"teacher":    option.Teacher,
					"time":       time.Now(),
				},
			},
		},
		bson.M{
			"$project": bson.M{
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
				Editable:   option.Editable,
				StudyPlace: entities.StudyPlace{},
				Group:      option.Subject,
				Teacher:    option.Group,
				Subject:    option.Teacher,
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

func (j *journalRepository) AddMarks(ctx context.Context, marks []entities.Mark, teacher string) error {
	if _, err := j.lessonsCollection.UpdateOne(ctx, bson.M{"_id": marks[0].LessonID, "teacher": teacher}, hMongo.PushArray("marks", marks)); err != nil {
		return err
	}

	return nil
}

func (j *journalRepository) AddMark(ctx context.Context, mark entities.Mark, teacher string) (primitive.ObjectID, error) {
	mark.Id = primitive.NewObjectID()
	if _, err := j.lessonsCollection.UpdateOne(ctx, bson.M{"_id": mark.LessonID, "teacher": teacher}, hMongo.Push("marks", mark)); err != nil {
		return primitive.NilObjectID, err
	}

	return mark.Id, nil
}

func (j *journalRepository) UpdateMark(ctx context.Context, mark entities.Mark, teacher string) error {
	if _, err := j.lessonsCollection.UpdateOne(ctx, bson.M{"_id": mark.LessonID, "teacher": teacher, "marks._id": mark.Id}, bson.M{"$set": bson.M{"marks.$": mark}}); err != nil {
		return err
	}

	return nil
}

func (j *journalRepository) DeleteMarkByID(ctx context.Context, id primitive.ObjectID, teacher string) error {
	if _, err := j.lessonsCollection.UpdateOne(ctx, bson.M{"teacher": teacher, "marks._id": id}, bson.M{"$pull": bson.M{"marks": bson.M{"_id": id}}}); err != nil {
		return err
	}

	return nil
}

func (j *journalRepository) AddAbsences(ctx context.Context, absences []entities.Absence, teacher string) error {
	if _, err := j.lessonsCollection.UpdateOne(ctx, bson.M{"_id": absences[0].LessonID, "teacher": teacher}, hMongo.PushArray("absences", absences)); err != nil {
		return err
	}

	return nil
}

func (j *journalRepository) AddAbsence(ctx context.Context, absence entities.Absence, teacher string) (primitive.ObjectID, error) {
	absence.Id = primitive.NewObjectID()
	if _, err := j.lessonsCollection.UpdateOne(ctx, bson.M{"_id": absence.LessonID, "teacher": teacher}, hMongo.Push("absences", absence)); err != nil {
		return primitive.NilObjectID, err
	}

	return absence.Id, nil
}

func (j *journalRepository) UpdateAbsence(ctx context.Context, absence entities.Absence, teacher string) error {
	if _, err := j.lessonsCollection.UpdateOne(ctx, bson.M{"_id": absence.LessonID, "teacher": teacher, "absences._id": absence.Id}, bson.M{"$set": bson.M{"absences.$": absence}}); err != nil {
		return err
	}

	return nil
}

func (j *journalRepository) DeleteAbsenceByID(ctx context.Context, id primitive.ObjectID, teacher string) error {
	if _, err := j.lessonsCollection.UpdateOne(ctx, bson.M{"teacher": teacher, "absences._id": id}, bson.M{"$pull": bson.M{"absences": bson.M{"_id": id}}}); err != nil {
		return err
	}

	return nil
}
