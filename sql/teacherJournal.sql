db.Users.aggregate([
    {
        "$group": {"_id": null, "users": {"$push": "$$ROOT"}}
    },
    {
        "$project": {
            "users._id": 1,
            "users.name": 1,
            "users.type": 1,
            "users.typename": 1
        }
    },
    {
        "$lookup": {
            "from": "SignUpCodes",
            "pipeline": [
                {
                    "$project": {
                        "name": 1,
                        "type": 1,
                        "typename": 1
                    }
                },
            ],
            "as": "codeUsers",
        }
    },
    {
        "$project": {
            "users": {
                "$filter": {
                    "input": {"$concatArrays": ["$codeUsers", "$users"]},
                    "as": "user",
                    "cond": {"$and": [{"$eq": ["$$user.type", "group"]}, {"$eq": ["$$user.typename", "95Т"]}]}
                }
            }
        }
    },
    {
        "$lookup": {
            "from": "Lessons",
            "let": {"userID": "$_id"},
            "pipeline": [
                {
                    "$match": {
                        "subject": "КПиЯП",
                        "group": "95Т"
                    }
                },
            ],
            "as": "lessons"
        }
    },
    {
        "$lookup": {
            "from": "StudyPlaces",
            "pipeline": [],
            "as": "studyPlace"
        }
    },
    {
        "$addFields": {
            "lessons": {
                "$sortArray": {
                    "input": "$lessons",
                    "sortBy": {"startDate": 1},
                }
            }
        }
    },
    {
        "$unwind": "$users"
    },
    {
        "$unwind": "$lessons"
    },
    {
        "$lookup": {
            "from": "Marks",
            "localField": "lessons._id",
            "foreignField": "lessonID",
            "let": {"userID": "$users._id"},
            "pipeline": [
                {
                    "$match": {
                        "$expr": {"$and": [{"$eq": ["$studentID", "$$userID"]}]}
                    }
                }
            ],
            "as": "lessons.marks"
        }
    },
    {
      "$addFields": {
          "lessons.journalCellColor": {
              "$function": {
                  // language=JavaScript
                  "body": `function (studyPlace, lesson) {
                        if (lesson === undefined || lesson.marks === undefined) return "";

                        let color = studyPlace.journalColors.general
                        for (let mark of lesson.marks) {
                            let type = studyPlace.lessonTypes.find(v => v.type === lesson.type);
                            if (type === undefined) return studyPlace.journalColors.general;

                            let markType = type.marks.find(m => m.mark === mark.mark);
                            if (markType === undefined || markType.workOutTime === undefined) return studyPlace.journalColors.general;

                            lesson.startDate.setSeconds(lesson.startDate.getSeconds() + markType.workOutTime);

                            color = lesson.startDate.getTime() > new Date().getTime() ? studyPlace.journalColors.warning : studyPlace.journalColors.danger;
                        }

                        return color;
                    }`,
                  "args": [{"$first": "$studyPlace"}, "$lessons"],
                  "lang": "js",
              },
          },
      }
    },
    {
        "$group": {
            "_id": {
                "_id": "$users._id",
                "title": "$users.name",
            },
            "lessons": {"$push": "$lessons"},
            "studyPlace": {"$first": {"$first": "$studyPlace"}}
        }
    },
    {
        "$project": {
            "row": {
                "_id": "$_id._id",
                "title": "$_id.title",
                "lessons": "$lessons"
            },
            "studyPlace": "$studyPlace",
        }
    },
    {
        "$group": {
            "_id": null,
            "dates": {"$first": "$row.lessons"},
            "rows": {"$push": "$row"},
            "studyPlace": {"$first": "$studyPlace"}
        }
    },
    {
        "$project": {
            "dates.marks": 0,
            "dates.journalCellColor": 0,
            "dates.studyPlace": 0,
            "rows.studyPlace": 0
        }
    },
    {
        "$addFields": {
            "info": {
                "editable": true,
                "studyPlace": "$studyPlace",
            }
        }
    },
    {
        "$project": {
            "studyPlace": 0
        }
    }
])
