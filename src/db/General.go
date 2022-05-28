package db

import (
	h "studyum/src/api"
	"studyum/src/models"
)

func GetStudyPlaces(studyPlaces *[]models.StudyPlace) *models.Error {
	studyPlacesCursor, err := StudyPlacesCollection.Find(nil, nil)
	if err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	if err := studyPlacesCursor.All(nil, &studyPlaces); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}
