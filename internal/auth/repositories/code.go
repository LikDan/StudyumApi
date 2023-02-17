package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/auth/entities"
)

type Code interface {
	GetUserByCodeAndDelete(ctx context.Context, code string) (entities.UserCodeData, error)
	GetAppData(ctx context.Context, code string) (map[string]any, error)
}

type code struct {
	codesCollection *mongo.Collection
}

func NewCode(codesCollection *mongo.Collection) Code {
	return &code{codesCollection: codesCollection}
}

func (r *code) GetUserByCodeAndDelete(ctx context.Context, code string) (codeData entities.UserCodeData, err error) {
	err = r.codesCollection.FindOneAndDelete(ctx, bson.M{"code": code}).Decode(&codeData)
	return
}

func (r *code) GetAppData(ctx context.Context, code string) (data map[string]any, err error) {
	raw, err := r.codesCollection.FindOne(ctx, bson.M{"code": code}).DecodeBytes()
	if err != nil {
		return nil, err
	}

	err = raw.Lookup("appData").Unmarshal(&data)
	return
}
