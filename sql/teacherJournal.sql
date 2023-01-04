db.Users.aggregate([
    {
        "$group": {"_id": null, "user": {"$push": "$$ROOT"}}
    },
    {
        "$project": {
            "user._id": 1,
            "user.name": 1,
            "user.type": 1,
            "user.typename": 1
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
            "user": {
                "$filter": {
                    "input": {"$concatArrays": ["$codeUsers", "$user"]},
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
        "$unwind": "$user"
    },
    {
        "$unwind": "$lessons"
    },
    {
        "$lookup": {
            "from": "Marks",
            "localField": "lessons._id",
            "foreignField": "lessonID",
            "let": {"userID": "$user._id"},
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
                "_id": "$user._id",
                "title": "$user.name",
            },
            "lessons": {"$push": "$lessons"},
            "studyPlace": {"$first": {"$first": "$studyPlace"}}
        }
    },
    {
      "$addFields": {
          "info": {
              "$function": {
                  "body": `function (studyPlace, lessons) {
							let info = {}

							let marks = lessons.flatMap(l => l?.marks ?? []).map(m => Number.parseInt(m.mark)).filter(m => m)
                            info.numericMarksSum = marks.reduce((sum, a) => sum + a, 0)
                            info.numericMarksAmount = marks.length

                            let color = studyPlace.journalColors.general
                            for (let lesson of lessons) {
                                if (lesson == null) continue

                                if (lesson.journalCellColor == studyPlace.journalColors.warning)
                                    color = studyPlace.journalColors.warning

                                if (lesson.journalCellColor == studyPlace.journalColors.danger){
                                    color = studyPlace.journalColors.danger
                                    break
                                }
                            }

                            info.color = color
							return info
                        }`,
                  "args": ["$studyPlace", "$lessons"],
                  "lang": "js",
              }
          }
      }
    },
    {
        "$project": {
            "row": {
                "_id": "$_id._id",
                "title": "$_id.title",
                "numericMarksSum": "$info.numericMarksSum",
                "numericMarksAmount": "$info.numericMarksAmount",
                "color": "$info.color",
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
