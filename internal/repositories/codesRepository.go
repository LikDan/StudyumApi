package repositories

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type CodesRepository struct {
	signUpCollection *mongo.Collection
}

func NewCodesRepository(client *mongo.Client) *CodesRepository {
	database := client.Database("Codes")

	return &CodesRepository{
		signUpCollection: database.Collection("SignUp"),
	}
}
