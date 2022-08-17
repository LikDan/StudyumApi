package controllers

import (
	"context"
	"studyum/src/models"
)

type IGeneralController interface {
	GetStudyPlaces(ctx context.Context) (error, []models.StudyPlace)
}
