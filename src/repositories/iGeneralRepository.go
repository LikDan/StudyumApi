package repositories

import (
	"context"
	"studyum/src/models"
)

type IGeneralRepository interface {
	GetAllStudyPlaces(ctx context.Context) (*models.Error, []models.StudyPlace)
}
