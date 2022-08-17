package controllers

import (
	"context"
	"studyum/src/entities"
)

type IGeneralController interface {
	GetStudyPlaces(ctx context.Context) (error, []entities.StudyPlace)
}
