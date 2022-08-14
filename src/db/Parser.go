package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
	"studyum/src/utils"
	"time"
)

func GetLessonByDate(date time.Time, name string, group string, lesson *models.Lesson) {
	result := lessonsCollection.FindOne(nil, bson.M{"subject": name, "group": group, "startDate": bson.M{"$gte": date, "$lt": date.AddDate(0, 0, 1)}})
	_ = result.Decode(lesson)
}

func GetUsersToParse(parserAppName string, users *[]models.ParseJournalUser) *models.Error {
	result, err := parseJournalUserCollection.Find(nil, bson.M{"parserAppName": parserAppName})
	if err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	if err := result.All(nil, users); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func InsertScheduleTypes(types []*models.ScheduleTypeInfo) *models.Error {
	if len(types) == 0 {
		return models.BindErrorStr("Provided empty array", 418, models.UNDEFINED)
	}

	if _, err := parseScheduleTypesCollection.DeleteMany(nil, bson.M{"parserAppName": types[0].ParserAppName}); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	for _, type_ := range types {
		type_.Id = primitive.NewObjectID()
	}

	if _, err := parseScheduleTypesCollection.InsertMany(nil, utils.ToInterfaceSlice(types)); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func GetScheduleTypesToParse(parserAppName string, types *[]models.ScheduleTypeInfo) *models.Error {
	result, err := parseScheduleTypesCollection.Find(nil, bson.M{"parserAppName": parserAppName})
	if err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	if err := result.All(nil, types); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func UpdateParseJournalUser(user *models.ParseJournalUser) *models.Error {
	if _, err := parseJournalUserCollection.UpdateByID(nil, user.ID, bson.M{"$set": user}); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func UpdateGeneralSchedule(lessons []*models.GeneralLesson) *models.Error {
	if len(lessons) == 0 {
		return models.BindErrorStr("Provided empty array", 418, models.UNDEFINED)
	}

	_, err := generalLessonsCollection.DeleteMany(nil, bson.D{{"studyPlaceId", lessons[0].StudyPlaceId}})
	if err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	_, err = generalLessonsCollection.InsertMany(nil, utils.ToInterfaceSlice(lessons))
	if err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}
