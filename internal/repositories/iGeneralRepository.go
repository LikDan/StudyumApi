package repositories

import (
	"context"
	"studyum/internal/entities"
)

type IGeneralRepository interface {
	GetAllStudyPlaces(ctx context.Context) (error, []entities.StudyPlace)
}
