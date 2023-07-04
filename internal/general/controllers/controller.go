package controllers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
	auth "studyum/internal/auth/entities"
	"studyum/internal/general/entities"
	"studyum/internal/general/repositories"
)

type Controller interface {
	GetStudyPlaces(ctx context.Context, restricted bool) (error, []entities.StudyPlace)
	GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (error, entities.StudyPlace)
	GetSelfStudyPlace(ctx context.Context, user auth.User) (error, entities.StudyPlace)
}

type controller struct {
	repository repositories.Repository
}

func NewGeneralController(repository repositories.Repository) Controller {
	return &controller{repository: repository}
}

func (g *controller) GetStudyPlaces(ctx context.Context, restricted bool) (error, []entities.StudyPlace) {
	return g.repository.GetAllStudyPlaces(ctx, restricted)
}

func (g *controller) GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (error, entities.StudyPlace) {
	return g.repository.GetStudyPlaceByID(ctx, id, restricted)
}

func (g *controller) GetSelfStudyPlace(ctx context.Context, user auth.User) (error, entities.StudyPlace) {
	return g.GetStudyPlaceByID(ctx, user.StudyPlaceInfo.ID, false)
}
