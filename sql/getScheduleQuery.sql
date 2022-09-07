let studyPlaceID = new ObjectId()
let type = "type"
let typename = "typename"
let startWeekDate = new Date()

db.StudyPlaces.aggregate(
    {
        $match: {
            "_id": studyPlaceID
        }
    },
    {
        $addFields: {
            "env": {
                "studyPlaceID": studyPlaceID,
                "startDate": startWeekDate,
                "endDate": {
                    $dateAdd: {
                        "startDate": startWeekDate,
                        "unit": "week",
                        "amount": "$weeksCount",
                    }
                },
                "weeksAmount": "$weeksCount"
            }
        }
    },
    {
        $lookup: {
            "from": "Lessons",
            "let": {"env": "$env"},
            "pipeline": [
                {
                    $match: {
                        $expr: {
                            $and: [
                                {
                                    $eq: ["$studyPlaceId", "$$env.studyPlaceID"],
                                }, {
                                    $eq: ['$' + type, typename],
                                }, {
                                    $gte: ["$startDate", "$$env.startDate"],
                                }, {
                                    $lt: ["$endDate", "$$env.endDate"],
                                }
                            ]
                        }
                    }
                },
                {
                    $addFields: {
                        "isGeneral": false
                    }
                }
            ],
            "as": "lessons"
        }
    },
    {
        $addFields: {
            "env.lastUpdatedDate": {$max: "$lessons.endDate"}
        }
    },
    {
        $addFields: {
            "env.startGeneral": {
                $dateFromParts: {
                    'year': {$year: "$env.lastUpdatedDate"},
                    'month': {$month: "$env.lastUpdatedDate"},
                    'day': {$sum: [{$dayOfMonth: "$env.lastUpdatedDate"}, 1]},
                }
            }
        }
    },
    {
        $addFields: {
            "env.startWeekIndex": {$mod: [{$isoWeek: "$env.startDate"}, "$env.weeksAmount"]},
            "env.startGeneralDayIndex": {$subtract: [{$isoDayOfWeek: "$env.startGeneral"}, 1]},
            "env.startGeneralWeekIndex": {$mod: [{$isoWeek: "$env.startGeneral"}, "$env.weeksAmount"]},
            "env.endGeneralDayIndex": {$subtract: [{$isoDayOfWeek: "$env.endDate"}, 1]},
            "env.endGeneralWeekIndex": {$mod: [{$isoWeek: "$env.endDate"}, "$env.weeksAmount"]},
        }
    },
    {
        $lookup: {
            "from": "GeneralLessons",
            "let": {"env": "$env"},
            "pipeline": [
                {
                    $match: {
                        $expr: {
                            $and: [
                                {
                                    $eq: ["$studyPlaceId", "$$env.studyPlaceID"],
                                }, {
                                    $eq: ['$' + type, typename],
                                },
                            ]
                        }
                    }
                },
                {
                    $addFields: {
                        "date": {
                            $dateAdd: {
                                "startDate": {
                                    $dateAdd: {
                                        "startDate": "$$env.startDate",
                                        "unit": "week",
                                        "amount": {$abs: {$subtract: ["$weekIndex", "$$env.startWeekIndex"]}}
                                    }
                                },
                                "unit": "day",
                                "amount": "$dayIndex"
                            }
                        }
                    }
                },
                {
                    $match: {
                        $expr: {
                            $and: [
                                {$gte: ["$date", "$$env.startGeneral"]},
                                {$lt: ["$date", "$$env.endDate"]}
                            ]
                        }
                    }
                },
                {
                    $addFields: {
                        "startDate": {
                            $toDate: {
                                $concat: [{
                                    $dateToString: {
                                        "format": "%Y-%m-%d",
                                        "date": "$date"
                                    }
                                }, "T", "$startTime"]
                            }
                        },
                        "endDate": {
                            $toDate: {
                                $concat: [{
                                    $dateToString: {
                                        "format": "%Y-%m-%d",
                                        "date": "$date"
                                    }
                                }, "T", "$endTime"]
                            }
                        },
                        "isGeneral": false
                    }
                },
            ],
            "as": "general"
        }
    },
    {
        $addFields: {
            "lessons": {$concatArrays: ["$lessons", "$general"]}
        }
    },
    {
        $addFields: {
            "_id": null,
            "studyPlace": "$$ROOT",
            "lessons": "$lessons"
        }
    },
    {
        $project: {
            "studyPlace.lessons": 0,
            "studyPlace.general": 0,
            "studyPlace.env": 0
        }
    },
    {
        $project: {
            "studyPlace": 1,
            "lessons": 1,
        }
    }
    )