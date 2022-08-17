package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"studyum/internal/entities"
)

type GeneralRepository struct {
	*Repository
}

func NewGeneralRepository(repository *Repository) *GeneralRepository {
	return &GeneralRepository{
		Repository: repository,
	}
}

func (g *GeneralRepository) GetAllStudyPlaces(ctx context.Context) (error, []entities.StudyPlace) {
	var studyPlaces []entities.StudyPlace
	studyPlacesCursor, err := g.studyPlacesCollection.Find(ctx, bson.M{})
	if err != nil {
		return err, nil
	}

	if err = studyPlacesCursor.All(ctx, &studyPlaces); err != nil {
		return err, nil
	}

	return nil, studyPlaces
}
