package kbp

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/apps/entities"
)

type app struct {
}

func NewApp() entities.App {
	return &app{}
}

func (a *app) GetStudyPlaceID(context.Context) primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex("631261e11b8b855cc75cec35")
	return id
}
