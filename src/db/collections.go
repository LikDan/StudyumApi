package db

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var GeneralSubjectsCollection *mongo.Collection
var SubjectsCollection *mongo.Collection
var StateCollection *mongo.Collection
var StudyPlacesCollection *mongo.Collection
var UsersCollection *mongo.Collection
var MarksCollection *mongo.Collection

func Init() {
	client, err := mongo.NewClient(options.Client().ApplyURI(getDbUrl()))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(nil)
	if err != nil {
		log.Fatalf("Can't connect to database, error: %s", err.Error())
		return
	}

	StudyPlacesCollection = client.Database("Schedule").Collection("StudyPlaces")

	UsersCollection = client.Database("Schedule").Collection("Users")

	SubjectsCollection = client.Database("Schedule").Collection("Subjects")
	GeneralSubjectsCollection = client.Database("Schedule").Collection("General")
	StateCollection = client.Database("Schedule").Collection("States")

	MarksCollection = client.Database("Schedule").Collection("Marks")
}
