package repositories

import (
	"context"
	"studyum/src/models"
)

type IGeneralRepository interface {
	GetStudyPlaces(ctx context.Context) (*models.Error, []*models.StudyPlace)
}
