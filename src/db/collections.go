package db

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

var (
	DataBase *mongo.Database

	GeneralLessonsCollection *mongo.Collection
	LessonsCollection        *mongo.Collection
	StateCollection          *mongo.Collection
	StudyPlacesCollection    *mongo.Collection
	UsersCollection          *mongo.Collection
	MarksCollection          *mongo.Collection

	ParseJournalUserCollection   *mongo.Collection
	ParseScheduleTypesCollection *mongo.Collection
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

	DataBase = client.Database("Schedule")

	StudyPlacesCollection = DataBase.Collection("StudyPlaces")
	UsersCollection = DataBase.Collection("Users")
	LessonsCollection = DataBase.Collection("Lessons")
	GeneralLessonsCollection = DataBase.Collection("GeneralLessons")
	StateCollection = DataBase.Collection("States")
	MarksCollection = DataBase.Collection("Marks")

	ParseJournalUserCollection = DataBase.Collection("ParseJournalUsers")
	ParseScheduleTypesCollection = DataBase.Collection("ParseScheduleTypes")
}
