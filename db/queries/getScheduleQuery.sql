db.GeneralLessons.aggregate([
    {"$match": {group: "95Т"}},
    {
        "$group": {
            "_id": {"dayIndex": "$dayIndex", "weekIndex": "$weekIndex"},
            "lessons": {"$push": "$$ROOT"},
        }
    },
    {
        "$sort": {
            "_id.weekIndex": 1,
            "_id.dayIndex": 1,
        }
    },
    {
        "$group": {
            "_id": null,
            "days": {"$push": "$$ROOT"},
        }
    },
    {
        "$addFields": {
            "days": {
                "$function": {
                    "body": `function(templates, start, end) {
    						const getWeekNumber = (date) => {
								const yearStart = new Date(date.getFullYear(), 0, 1);
								return Math.ceil((((date.getTime() - yearStart.getTime()) / 86400000) + yearStart.getDay() + 1) / 7);
    						}

    						const weekAmount = Math.max(...templates.map(t => t._id.weekIndex)) + 1
							const currentDate = new Date(start.getTime());
							const lessons = [];
							while (currentDate <= end) {
								const day = currentDate.getUTCDay() === 0 ? 6 : currentDate.getUTCDay() - 1;
								const week = getWeekNumber(currentDate) % weekAmount;
                                const template = templates.find(t => t._id.dayIndex === day && t._id.weekIndex === week)
                                if (!!template) {
                                    template.lessons = template.lessons.map(t => {
                                        const date = new Date(currentDate.getTime())
                                        const startDate = new Date(currentDate.toLocaleDateString() + ' ' + t.startTime)
                                        const endDate = new Date(currentDate.toLocaleDateString() + ' ' + t.endTime)
                                        return {...t, date: day, startDate, endDate, isGeneral: true}
                                    })
                                    lessons.push({...template});
                                }
								currentDate.setDate(currentDate.getDate() + 1);
							}

							return lessons
						}`,
                    "args": ["$days", new Date(2023, 4, 17), new Date(2023, 4, 30)],
                    "lang": "js",
                },
            },
        },
    },
    {"$unwind": "$days"},
    {"$replaceRoot": {"newRoot": "$days"}},
    {"$addFields": {"_id": {"$first": "$lessons.date"}}},
    {"$project": {"general": "$lessons"}},
    {
        "$lookup": {
            "from": "Lessons",
            "let": {"from": "$_id.date", "till": {"$dateAdd": {"startDate": "$_id.date", "unit": "day", "amount": 1}}},
            "pipeline": [
                {
                    "$match": {
                        "$expr": {
                            "$and": [{"$eq": ["$group", "95Т"]}, {"$gte": ["$startDate", "$$from"]}, {"$lt": ["$startDate", "$$till"]}]
                        }
                    }
                },
            ],
            "as": "lessons"
        }
    },
    {
        "$project": {
            "lessons": {
                "$cond": {
                    "if": {"$eq": ["$lessons", []]},
                    "then": "$general",
                    "else": "$lessons"
                }
            }
        }
    },
    {"$unwind": "$lessons"},
    {"$replaceRoot": {"newRoot": "$lessons"}},
    {
        "$group": {
            "_id": null,
            "lessons": {"$push": "$$ROOT"}
        }
    },
])