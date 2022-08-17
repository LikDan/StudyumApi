package controllers

import (
	"context"
	"studyum/internal/entities"
)

type IGeneralController interface {
	GetStudyPlaces(ctx context.Context) (error, []entities.StudyPlace)
}
