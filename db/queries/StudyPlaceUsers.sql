db.createView('StudyPlaceUsers', 'Users', [
        {
          "$addFields": {
            "studyPlaceInfo.studyPlaceID": "$studyPlaceInfo._id",
            "studyPlaceInfo._id": "$_id"
          }
        },
        {
          "$replaceRoot": {
            "newRoot": "$studyPlaceInfo"
          }
        },
        {
          "$group": {
            "_id": null,
            "user": {
              "$push": "$$ROOT"
            }
          }
        },
        {
          "$lookup": {
            "from": "CodeUsers",
            "pipeline": [],
            "as": "codeUsers"
          }
        },
        {
          "$project": {
            "users": {
              "$concatArrays": ["$codeUsers", "$user"]
            }
          }
        },
        {
          "$unwind": "$users"
        },
        {
          "$replaceRoot": {
            "newRoot": "$users"
          }
        },
        {
          "$project": {
            "_id": 1,
            "studyPlaceID": 1,
            "name": 1,
            "role": 1,
            "roleName": 1,
            "tuitionGroup": 1,
            "permissions": 1
          }
        }
      ])