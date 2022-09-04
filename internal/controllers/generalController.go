package controllers

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
	"studyum/internal/repositories"
)

type GeneralController interface {
	GetStudyPlaces(ctx context.Context, restricted bool) (error, []entities.StudyPlace)
	GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (error, entities.StudyPlace)
}

type generalController struct {
	repository repositories.GeneralRepository
}

func NewGeneralController(repository repositories.GeneralRepository) GeneralController {
	return &generalController{repository: repository}
}

func (g *generalController) GetStudyPlaces(ctx context.Context, restricted bool) (error, []entities.StudyPlace) {
	return g.repository.GetAllStudyPlaces(ctx, restricted)
}

func (g *generalController) GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (error, entities.StudyPlace) {
	return g.repository.GetStudyPlaceByID(ctx, id, restricted)
}
