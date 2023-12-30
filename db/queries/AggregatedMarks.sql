db.createView('AggregatedMarks', 'Marks', [
    {
        "$project": {
            "root": "$$ROOT"
        }
    },
    {
        "$lookup": {
            "from": "StudyPlaceMarks",
            "localField": "root.markID",
            "foreignField": "_id",
            "as": "mark",
        }
    },
    {
        "$replaceRoot": {
            "newRoot": {
                "$arrayToObject": {
                    "$map": {
                        "input": {
                            "$concatArrays": [
                                {"$objectToArray": {"$first": "$mark"}},
                                {"$objectToArray": "$root"},
                            ]
                        },
                        "as": "field",
                        "in": {
                            "k": "$$field.k",
                            "v": "$$field.v",
                        },
                    },
                },
            },
        },
    },
    {
        "$project": {
            "assignLessonTypeIDs": 0
        }
    }
]);
