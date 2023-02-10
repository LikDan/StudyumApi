package repositories

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"studyum/internal/auth/entities"
	entities2 "studyum/internal/general/entities"
)

type Middleware interface {
	GetUserByID(ctx context.Context, id primitive.ObjectID) (entities.User, error)

	GetStudyPlaceByApiToken(ctx context.Context, token string) (entities2.StudyPlace, error)
}

type middleware struct {
	users       *mongo.Collection
	studyPlaces *mongo.Collection
}

func NewMiddleware(users *mongo.Collection, studyPlaces *mongo.Collection) Middleware {
	return &middleware{users: users, studyPlaces: studyPlaces}
}

func (r *middleware) GetUserByID(ctx context.Context, id primitive.ObjectID) (user entities.User, err error) {
	err = r.users.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return
}

func (r *middleware) GetStudyPlaceByApiToken(ctx context.Context, token string) (studyPlace entities2.StudyPlace, err error) {
	err = r.studyPlaces.FindOne(ctx, bson.M{"apiToken": token}).Decode(&studyPlace)
	return
}
