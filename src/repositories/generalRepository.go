package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"studyum/src/models"
)

type GeneralRepository struct {
	*Repository
}

func NewGeneralRepository(repository *Repository) *GeneralRepository {
	return &GeneralRepository{
		Repository: repository,
	}
}

func (g *GeneralRepository) GetStudyPlaces(ctx context.Context) (*models.Error, []*models.StudyPlace) {
	var studyPlaces []*models.StudyPlace
	studyPlacesCursor, err := g.studyPlacesCollection.Find(ctx, bson.M{})
	if err != nil {
		return models.BindError(err, 418, models.WARNING), nil
	}

	if err := studyPlacesCursor.All(ctx, &studyPlaces); err != nil {
		return models.BindError(err, 418, models.WARNING), nil
	}

	return models.EmptyError(), studyPlaces
}
