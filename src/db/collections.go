package db

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

var (
	database *mongo.Database

	generalLessonsCollection *mongo.Collection
	lessonsCollection        *mongo.Collection
	studyPlacesCollection    *mongo.Collection
	usersCollection          *mongo.Collection
	marksCollection          *mongo.Collection

	parseJournalUserCollection   *mongo.Collection
	parseScheduleTypesCollection *mongo.Collection
)

func Init() {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL")))
	if err != nil {
		log.Fatal(err)
	}

	if err = client.Connect(nil); err != nil {
		log.Fatalf("Can't connect to database, error: %s", err.Error())
		return
	}

	database = client.Database("Schedule")

	studyPlacesCollection = database.Collection("StudyPlaces")
	usersCollection = database.Collection("Users")
	lessonsCollection = database.Collection("Lessons")
	generalLessonsCollection = database.Collection("GeneralLessons")
	marksCollection = database.Collection("Marks")

	parseJournalUserCollection = database.Collection("ParseJournalUsers")
	parseScheduleTypesCollection = database.Collection("ParseScheduleTypes")
}
