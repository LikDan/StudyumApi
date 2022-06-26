package db

import (
	h "studyum/src/api"
	"studyum/src/models"
)

func AddMark(mark *models.Mark) *models.Error {
	if _, err := MarksCollection.InsertOne(nil, mark); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}

func AddMarks(marks []*models.Mark) *models.Error {
	if _, err := MarksCollection.InsertMany(nil, h.ToInterfaceSlice(marks)); err != nil {
		return models.BindError(err, 418, h.WARNING)
	}

	return models.EmptyError()
}
