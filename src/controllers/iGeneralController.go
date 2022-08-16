package controllers

import (
	"context"
	"studyum/src/models"
)

type IGeneralController interface {
	GetStudyPlaces(ctx context.Context) (*models.Error, []models.StudyPlace)
}
