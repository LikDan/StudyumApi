package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/general/entities"
)

type Repository interface {
	GetAllStudyPlaces(ctx context.Context, restricted bool) (error, []entities.StudyPlace)
	GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (error, entities.StudyPlace)
	GetStudyPlaceByApiToken(ctx context.Context, token string) (error, entities.StudyPlace)
}

type repository struct {
	studyPlaces *mongo.Collection
}

func NewGeneralRepository(studyPlaces *mongo.Collection) Repository {
	return &repository{studyPlaces: studyPlaces}
}

func (g *repository) GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (err error, studyPlace entities.StudyPlace) {
	err = g.studyPlaces.FindOne(ctx, bson.M{"_id": id, "restricted": restricted}).Decode(&studyPlace)
	return
}

func (g *repository) GetAllStudyPlaces(ctx context.Context, restricted bool) (error, []entities.StudyPlace) {
	filter := bson.M{}
	if !restricted {
		filter["restricted"] = false
	}

	var studyPlaces []entities.StudyPlace
	studyPlacesCursor, err := g.studyPlaces.Find(ctx, filter)
	if err != nil {
		return err, nil
	}

	if err = studyPlacesCursor.All(ctx, &studyPlaces); err != nil {
		return err, nil
	}

	return nil, studyPlaces
}

func (g *repository) GetStudyPlaceByApiToken(ctx context.Context, token string) (err error, studyPlace entities.StudyPlace) {
	err = g.studyPlaces.FindOne(ctx, bson.M{"apiToken": token}).Decode(&studyPlace)
	return
}
