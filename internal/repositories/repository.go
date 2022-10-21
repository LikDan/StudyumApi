package repositories

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	generalLessonsCollection *mongo.Collection
	lessonsCollection        *mongo.Collection
	studyPlacesCollection    *mongo.Collection
	usersCollection          *mongo.Collection
	absencesCollection       *mongo.Collection
	signUpCodesCollection    *mongo.Collection
}

func NewRepository(client *mongo.Client) *Repository {
	database := client.Database("Studyum")

	return &Repository{
		generalLessonsCollection: database.Collection("GeneralLessons"),
		lessonsCollection:        database.Collection("Lessons"),
		studyPlacesCollection:    database.Collection("StudyPlaces"),
		usersCollection:          database.Collection("Users"),
		absencesCollection:       database.Collection("Absences"),
		signUpCodesCollection:    database.Collection("SignUpCodes"),
	}
}
