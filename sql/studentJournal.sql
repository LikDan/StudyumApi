db.Lessons.aggregate([
    {$match: {group: "95Т"}}, //TODO studyPlaceID
    {
        $lookup: {
            "from": "Marks",
            "localField": "_id",
            "foreignField": "lessonID",
            "pipeline": [], //TODO studyPlaceID and userID
            "as": "marks",
        }
    },
    {
        $lookup: {
            "from": "StudyPlaces",
            "pipeline": [], //TODO studyPlaceID
            "as": "studyPlace",
        }
    },
    {
        $addFields: {
            journalCellColor: {
                $function: {
                    // language=JavaScript
                    body: `function (studyPlace, lesson) {
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
                    args: [{$first: "$studyPlace"}, "$$ROOT"],
                    lang: "js",
                }
            },
        }
    },
    {
        $group: {
            _id: null,
            studyPlace: {$first: {$first: "$studyPlace"}},
            lessons: {$push: "$$ROOT"},
            dates: {$addToSet: {$toDate: {$dateToString: {date: "$startDate", format: "%m/%d/%Y"}}}}
        }
    },
    {
        $addFields: {
            "dates": {
                $sortArray: {
                    input: "$dates",
                    sortBy: 1
                }
            },
        }
    },
    {
        $project: {
            "lessons.studyPlace": 0,
        }
    },
    {
        $addFields: {
            rows: {
                $function: {
                    // language=JavaScript
                    body: `function (studyPlace, lessons, dates) {
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

                        return rows.sort((a, b) => a.title > b.title)
                    }`,
                    args: ["$studyPlace", "$lessons", "$dates"],
                    lang: "js",
                }
            }
        }
    },
    {
        $addFields: {
            "info": {
                "editable": false,
                "studyPlace": "$studyPlace"
            }
        }
    },
    {
        $addFields: {
            dates: {
                $map:
                    {
                        input: "$dates",
                        as: "date",
                        in: {
                            startDate: "$$date",
                            endDate: "$$date"
                        }
                    }
            }
        }
    },
    {
        $project: {
            lessons: 0,
            studyPlace: 0
        }
    },
])


db.StudyPlaces.updateOne({}, {$set: {lessonTypes: [{"type": "Лекция", "marks": [{"mark": "1", "workOutTime": 604800}, {"mark": "2", "workOutTime": 604800}, {"mark": "3"}, {"mark": "4"}, {"mark": "5"}, {"mark": "6"}, {"mark": "7"}, {"mark": "8"}, {"mark": "9"}, {"mark": "10"}]}, {"type": "Практика", "marks": [{"mark": "1", "workOutTime": 604800}, {"mark": "2", "workOutTime": 604800}, {"mark": "3"}, {"mark": "4"}, {"mark": "5"}, {"mark": "6"}, {"mark": "7"}, {"mark": "8"}, {"mark": "9"}, {"mark": "10"}]}, {"type": "Лабораторная", "marks": [{"mark": "1", "workOutTime": 604800}, {"mark": "2", "workOutTime": 604800}, {"mark": "3"}, {"mark": "4"}, {"mark": "5"}, {"mark": "6"}, {"mark": "7"}, {"mark": "8"}, {"mark": "9"}, {"mark": "10"}], "standaloneMarks": [{"mark": "зч", "workOutTime": 604800}]}]}})