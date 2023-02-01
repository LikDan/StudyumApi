package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/auth/entities"
)

type Auth interface {
	GetUserByLogin(ctx context.Context, login string) (entities.User, error)
	AddUser(ctx context.Context, user entities.User) error
	UpdateUser(ctx context.Context, user entities.User) error
	VerifyEmail(ctx context.Context, userID primitive.ObjectID) error
}

type auth struct {
	users *mongo.Collection
}

func NewAuth(usersCollection *mongo.Collection) Auth {
	return &auth{users: usersCollection}
}

func (r *auth) GetUserByLogin(ctx context.Context, login string) (user entities.User, err error) {
	err = r.users.FindOne(ctx, bson.M{"login": login}).Decode(&user)
	return
}

func (r *auth) AddUser(ctx context.Context, user entities.User) error {
	if user.Sessions == nil {
		user.Sessions = make([]entities.Session, 0, 1)
	}

	_, err := r.users.InsertOne(ctx, user)
	return err
}

func (r *auth) UpdateUser(ctx context.Context, user entities.User) error {
	_, err := r.users.UpdateOne(ctx, bson.M{"_id": user.Id}, bson.M{"$set": user})
	return err
}

func (r *auth) VerifyEmail(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.users.UpdateByID(ctx, userID, bson.M{"$set": bson.M{"verifiedEmail": true}})
	return err
}
