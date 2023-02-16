package schedule

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository interface {
	GetSubjectID(ctx context.Context, subject string) (string, error)
	GetGroupID(ctx context.Context, group string) (string, error)
}

type mongoRepository struct {
	db *mongo.Database
}

func NewMongoRepository(db *mongo.Database) MongoRepository {
	return &mongoRepository{db: db}
}

func (m *mongoRepository) getID(ctx context.Context, collection, nameVal string) (string, error) {
	val, err := m.db.Collection(collection).FindOne(ctx, bson.M{"name": nameVal}).DecodeBytes()
	if err != nil {
		return "", err
	}

	return val.Lookup("_id").StringValue(), nil
}

func (m *mongoRepository) GetSubjectID(ctx context.Context, subject string) (string, error) {
	return m.getID(ctx, "Subjects", subject)
}

func (m *mongoRepository) GetGroupID(ctx context.Context, group string) (string, error) {
	return m.getID(ctx, "Groups", group)
}
