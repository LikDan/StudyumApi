package controllers

import (
	"context"
	"studyum/internal/entities"
	"studyum/internal/repositories"
)

type GeneralController interface {
	GetStudyPlaces(ctx context.Context) (error, []entities.StudyPlace)
}

type generalController struct {
	repository repositories.GeneralRepository
}

func NewGeneralController(repository repositories.GeneralRepository) GeneralController {
	return &generalController{repository: repository}
}

func (g *generalController) GetStudyPlaces(ctx context.Context) (error, []entities.StudyPlace) {
	return g.repository.GetAllStudyPlaces(ctx)
}
