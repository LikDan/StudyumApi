db.Lessons.aggregate([
    {
        "$match": {"_id": new ObjectId("63b5275b5c953b713cdc425e")},
    },
    {
        "$lookup": {
            "from": "Lessons",
            "let": {
                "date": {
                    "$dateToString": {
                        "date": "$startDate",
                        "format": "%Y-%m-%d",
                    }
                },
                "subject": "$subject",
                "teacher": "$teacher",
                "group": "$group"
            },
            "pipeline": [
                {
                    "$match": {
                        "$expr": {
                            "$and": [
                                {
                                    "$eq": [
                                        "$$date", {
                                            "$dateToString": {
                                                "date": "$startDate",
                                                "format": "%Y-%m-%d",
                                            }
                                        },
                                    ]
                                },
                                {
                                    "$eq": ["$subject", "$$subject"]
                                },
                                {
                                    "$eq": ["$teacher", "$$teacher"]
                                },
                                {
                                    "$eq": ["$group", "$$group"]
                                }
                            ]
                        }
                    }
                },
                {
                    "$addFields": {
                        "marks": {
                            "$filter":
                                {
                                    "input": "$marks",
                                    "cond": {"$eq": ["$$marks.studentID", new ObjectId("633322073d379e063d8ee8a4")]},
                                    "as": "marks",
                                }
                        }
                    }
                }
            ],
            "as": "lessons",
        },
    },
    {
        "$project": {
            "lessons": 1,
        },
    },
    {
        "$unwind": "$lessons",
    },
    {
        "$replaceRoot": {
            "newRoot": "$lessons"
        }
    },
    {
        "$sort": {
            "startDate": 1
        }
    }
])