package repositories

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	database *mongo.Database

	generalLessonsCollection *mongo.Collection
	lessonsCollection        *mongo.Collection
	studyPlacesCollection    *mongo.Collection
	usersCollection          *mongo.Collection
	marksCollection          *mongo.Collection
}

func NewRepository(client *mongo.Client) *Repository {
	database := client.Database("Schedule")

	return &Repository{
		database:                 database,
		generalLessonsCollection: database.Collection("GeneralLessons"),
		lessonsCollection:        database.Collection("Lessons"),
		studyPlacesCollection:    database.Collection("StudyPlaces"),
		usersCollection:          database.Collection("Users"),
		marksCollection:          database.Collection("Marks"),
	}
}
