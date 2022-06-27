package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	h "studyum/src/api"
	"studyum/src/models"
	"time"
)

func GetLessonByDate(date time.Time, name string, group string, lesson *models.Lesson) {
	result := LessonsCollection.FindOne(nil, bson.M{"subject": name, "group": group, "startDate": bson.M{"$gte": date, "$lt": date.AddDate(0, 0, 1)}})
	_ = result.Decode(lesson)
}

func GetUsersToParse(parserAppName string, users *[]models.ParseJournalUser) *models.Error {
	result, err := ParseJournalUserCollection.Find(nil, bson.M{"parserAppName": parserAppName})
	if err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	if err := result.All(nil, users); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}

func InsertScheduleTypes(types []*models.ScheduleTypeInfo) *models.Error {
	if _, err := ParseScheduleTypesCollection.DeleteMany(nil, bson.M{"parserAppName": types[0].ParserAppName}); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	for _, type_ := range types {
		type_.Id = primitive.NewObjectID()
	}

	if _, err := ParseScheduleTypesCollection.InsertMany(nil, h.ToInterfaceSlice(types)); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}

func GetScheduleTypesToParse(parserAppName string, types *[]models.ScheduleTypeInfo) *models.Error {
	result, err := ParseScheduleTypesCollection.Find(nil, bson.M{"parserAppName": parserAppName})
	if err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	if err := result.All(nil, types); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}

func UpdateParseJournalUser(user *models.ParseJournalUser) *models.Error {
	if _, err := ParseJournalUserCollection.UpdateByID(nil, user.ID, bson.M{"$set": user}); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}

func UpdateGeneralSchedule(lessons []*models.GeneralLesson) *models.Error {
	_, err := GeneralLessonsCollection.DeleteMany(nil, bson.D{{"studyPlaceId", lessons[0].StudyPlaceId}})
	if err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	_, err = GeneralLessonsCollection.InsertMany(nil, h.ToInterfaceSlice(lessons))
	if err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}
