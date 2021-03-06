package db

import (
	"go.mongodb.org/mongo-driver/bson"
	h "studyum/src/api"
	"studyum/src/models"
)

func GetStudyPlaces() (*models.Error, []*models.StudyPlace) {
	var studyPlaces []*models.StudyPlace
	studyPlacesCursor, err := StudyPlacesCollection.Find(nil, bson.M{})
	if err != nil {
		return models.BindError(err, 418, h.WARNING), nil
	}

	if err := studyPlacesCursor.All(nil, &studyPlaces); err != nil {
		return models.BindError(err, 418, h.WARNING), nil
	}

	return models.EmptyError(), studyPlaces
}
