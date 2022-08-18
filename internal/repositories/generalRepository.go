package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"studyum/internal/entities"
)

type GeneralRepository interface {
	GetAllStudyPlaces(ctx context.Context) (error, []entities.StudyPlace)
}

type generalRepository struct {
	*Repository
}

func NewGeneralRepository(repository *Repository) GeneralRepository {
	return &generalRepository{Repository: repository}
}

func (g *generalRepository) GetAllStudyPlaces(ctx context.Context) (error, []entities.StudyPlace) {
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
