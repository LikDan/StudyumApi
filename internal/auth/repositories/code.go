package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/auth/entities"
)

type Code interface {
	GetUserByCodeAndDelete(ctx context.Context, code string) (entities.UserCodeData, error)
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
