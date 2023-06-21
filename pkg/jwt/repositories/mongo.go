package repositories

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	m "go.mongodb.org/mongo-driver/mongo"
	"studyum/pkg/jwt/entities"
	"time"
)

type mongo struct {
	sessions *m.Collection
}

func NewMongo(sessions *m.Collection) Repository {
	return &mongo{sessions: sessions}
}

func (r *mongo) Add(ctx context.Context, session entities.Session) error {
	_, err := r.sessions.InsertOne(ctx, session)
	return err
}

func (r *mongo) RemoveByID(ctx context.Context, id string) error {
	_, err := r.sessions.DeleteMany(ctx, bson.M{"_id": id})
	return err
}

func (r *mongo) GetByID(ctx context.Context, id string) (session entities.Session, err error) {
	err = r.sessions.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if errors.Is(err, m.ErrNoDocuments) {
		return entities.Session{}, NotValidRefreshTokenErr
	}
	return
}

func (r *mongo) Update(ctx context.Context, session entities.Session) error {
	_, err := r.sessions.UpdateByID(ctx, session.ID, bson.M{"$set": session})
	return err
}

func (r *mongo) RemoveExpired(ctx context.Context) (int, error) {
	many, err := r.sessions.DeleteMany(ctx, bson.M{"expire": bson.M{"$gte": time.Now()}})
	if err != nil {
		return 0, err
	}
	return int(many.DeletedCount), nil
}
