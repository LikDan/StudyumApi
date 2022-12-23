package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
)

type GeneralRepository interface {
	GetAllStudyPlaces(ctx context.Context, restricted bool) (error, []entities.StudyPlace)
	GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (error, entities.StudyPlace)
	GetStudyPlaceByApiToken(ctx context.Context, token string) (error, entities.StudyPlace)
}

type generalRepository struct {
	*Repository
}

func NewGeneralRepository(repository *Repository) GeneralRepository {
	return &generalRepository{Repository: repository}
}

func (g *generalRepository) GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (err error, studyPlace entities.StudyPlace) {
	err = g.studyPlacesCollection.FindOne(ctx, bson.M{"_id": id, "restricted": restricted}).Decode(&studyPlace)
	return
}

func (g *generalRepository) GetAllStudyPlaces(ctx context.Context, restricted bool) (error, []entities.StudyPlace) {
	filter := bson.M{}
	if !restricted {
		filter["restricted"] = false
	}

	var studyPlaces []entities.StudyPlace
	studyPlacesCursor, err := g.studyPlacesCollection.Find(ctx, filter)
	if err != nil {
		return err, nil
	}

	if err = studyPlacesCursor.All(ctx, &studyPlaces); err != nil {
		return err, nil
	}

	return nil, studyPlaces
}

func (g *generalRepository) GetStudyPlaceByApiToken(ctx context.Context, token string) (err error, studyPlace entities.StudyPlace) {
	err = g.studyPlacesCollection.FindOne(ctx, bson.M{"apiToken": token}).Decode(&studyPlace)
	return
}
