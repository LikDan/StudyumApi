package repositories

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"studyum/internal/auth/entities"
)

type Middleware interface {
	AddSession(ctx context.Context, session entities.Session, userID primitive.ObjectID) error
	DeleteSessionByToken(ctx context.Context, token string) error

	GetUserByID(ctx context.Context, id primitive.ObjectID) (entities.User, error)
}

type middleware struct {
	usersCollection *mongo.Collection
}

func NewMiddleware(usersCollection *mongo.Collection) Middleware {
	return &middleware{usersCollection: usersCollection}
}

func (r *middleware) AddSession(ctx context.Context, session entities.Session, userID primitive.ObjectID) error {
	_, err := r.usersCollection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$push": bson.M{"sessions": session}})
	return err
}

func (r *middleware) DeleteSessionByToken(ctx context.Context, token string) error {
	_, err := r.usersCollection.UpdateOne(ctx, bson.M{"sessions.token": token},
		bson.M{"$pull": bson.M{"sessions": bson.M{"token": token}}},
	)
	return err
}

func (r *middleware) GetUserByID(ctx context.Context, id primitive.ObjectID) (user entities.User, err error) {
	err = r.usersCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return
}
