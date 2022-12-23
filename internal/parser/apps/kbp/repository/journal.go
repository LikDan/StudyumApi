package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/parser/apps/kbp/entities"
)

type Repository struct {
	journalUsersCollection *mongo.Collection
}

func NewRepository(client *mongo.Client) *Repository {
	db := client.Database("Kbp")
	return &Repository{journalUsersCollection: db.Collection("JournalUsers")}
}

func (r *Repository) GetJournalUserByID(ctx context.Context, id string) (err error, user entities.JournalUser) {
	err = r.journalUsersCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return
}
