package general

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
)

type Controller interface {
	GetStudyPlaces(ctx context.Context, restricted bool) (error, []StudyPlace)
	GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (error, StudyPlace)
}

type controller struct {
	repository Repository
}

func NewGeneralController(repository Repository) Controller {
	return &controller{repository: repository}
}

func (g *controller) GetStudyPlaces(ctx context.Context, restricted bool) (error, []StudyPlace) {
	return g.repository.GetAllStudyPlaces(ctx, restricted)
}

func (g *controller) GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (error, StudyPlace) {
	return g.repository.GetStudyPlaceByID(ctx, id, restricted)
}
