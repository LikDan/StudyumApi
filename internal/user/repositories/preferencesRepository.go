package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"studyum/internal/user/entities"
)

type PreferencesRepository interface {
	GetPreferences(ctx context.Context, userID primitive.ObjectID) (entities.Preferences, error)
	SavePreferences(ctx context.Context, userID primitive.ObjectID, preferences entities.Preferences) error
}

type preferencesRepository struct {
	preferences *mongo.Collection
}

func NewPreferencesRepository(preferences *mongo.Collection) PreferencesRepository {
	return &preferencesRepository{preferences: preferences}
}

func (p *preferencesRepository) GetPreferences(ctx context.Context, userID primitive.ObjectID) (preferences entities.Preferences, err error) {
	err = p.preferences.FindOne(ctx, bson.M{"userID": userID}).Decode(&preferences)
	return
}

func (p *preferencesRepository) SavePreferences(ctx context.Context, userID primitive.ObjectID, preferences entities.Preferences) error {
	upsert := true
	_, err := p.preferences.UpdateOne(ctx, bson.M{"userID": userID}, bson.M{"$set": preferences}, &options.UpdateOptions{
		Upsert: &upsert,
	})
	return err
}
