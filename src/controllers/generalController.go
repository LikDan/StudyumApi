package controllers

import (
	"context"
	"studyum/src/models"
	"studyum/src/repositories"
)

type GeneralController struct {
	repository repositories.IGeneralRepository
}

func NewGeneralController(repository repositories.IGeneralRepository) *GeneralController {
	return &GeneralController{repository: repository}
}

func (g *GeneralController) GetStudyPlaces(ctx context.Context) (*models.Error, []models.StudyPlace) {
	return g.repository.GetAllStudyPlaces(ctx)
}
