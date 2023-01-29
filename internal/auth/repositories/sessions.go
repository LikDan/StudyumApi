package repositories

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"studyum/internal/auth/entities"
)

type Sessions interface {
	Add(ctx context.Context, session entities.Session, userID primitive.ObjectID) error
	RemoveByToken(ctx context.Context, token string) error
	GetUserByToken(ctx context.Context, token string) (entities.User, error)
}

type sessions struct {
	usersCollection *mongo.Collection
}

func NewSessions(usersCollection *mongo.Collection) Sessions {
	return &sessions{usersCollection: usersCollection}
}

func (r *sessions) Add(ctx context.Context, session entities.Session, userID primitive.ObjectID) error {
	_, err := r.usersCollection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$push": bson.M{"sessions": session}})
	return err
}

func (r *sessions) RemoveByToken(ctx context.Context, token string) error {
	_, err := r.usersCollection.UpdateOne(ctx, bson.M{"sessions.token": token},
		bson.M{"$pull": bson.M{"sessions": bson.M{"token": token}}},
	)
	return err
}

func (r *sessions) GetUserByToken(ctx context.Context, token string) (user entities.User, err error) {
	err = r.usersCollection.FindOne(ctx, bson.M{"sessions.token": token}).Decode(&user)
	return
}
