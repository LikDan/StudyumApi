package controllers

import (
	"context"
	"studyum/internal/entities"
	"studyum/internal/repositories"
)

type GeneralController struct {
	repository repositories.GeneralRepository
}

func NewGeneralController(repository repositories.GeneralRepository) *GeneralController {
	return &GeneralController{repository: repository}
}

func (g *GeneralController) GetStudyPlaces(ctx context.Context) (error, []entities.StudyPlace) {
	return g.repository.GetAllStudyPlaces(ctx)
}
