package general

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/global"
)

type Repository interface {
	GetAllStudyPlaces(ctx context.Context, restricted bool) (error, []StudyPlace)
	GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (error, StudyPlace)
	GetStudyPlaceByApiToken(ctx context.Context, token string) (error, StudyPlace)
}

type repository struct {
	*global.Repository
}

func NewGeneralRepository(r *global.Repository) Repository {
	return &repository{Repository: r}
}

func (g *repository) GetStudyPlaceByID(ctx context.Context, id primitive.ObjectID, restricted bool) (err error, studyPlace StudyPlace) {
	err = g.StudyPlacesCollection.FindOne(ctx, bson.M{"_id": id, "restricted": restricted}).Decode(&studyPlace)
	return
}

func (g *repository) GetAllStudyPlaces(ctx context.Context, restricted bool) (error, []StudyPlace) {
	filter := bson.M{}
	if !restricted {
		filter["restricted"] = false
	}

	var studyPlaces []StudyPlace
	studyPlacesCursor, err := g.StudyPlacesCollection.Find(ctx, filter)
	if err != nil {
		return err, nil
	}

	if err = studyPlacesCursor.All(ctx, &studyPlaces); err != nil {
		return err, nil
	}

	return nil, studyPlaces
}

func (g *repository) GetStudyPlaceByApiToken(ctx context.Context, token string) (err error, studyPlace StudyPlace) {
	err = g.StudyPlacesCollection.FindOne(ctx, bson.M{"apiToken": token}).Decode(&studyPlace)
	return
}
