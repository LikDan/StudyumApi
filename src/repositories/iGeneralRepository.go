package repositories

import (
	"context"
	"studyum/src/entities"
)

type IGeneralRepository interface {
	GetAllStudyPlaces(ctx context.Context) (error, []entities.StudyPlace)
}
