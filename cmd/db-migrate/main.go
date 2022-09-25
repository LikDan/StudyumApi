package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

func main() {
	ctx := context.Background()

	clientFrom, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL_FROM")))
	if err != nil {
		logrus.Fatal(err)
	}

	if err = clientFrom.Connect(ctx); err != nil {
		logrus.Fatalf("Can't connect to database, error: %s", err.Error())
	}

	clientTo, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL_TO")))
	if err != nil {
		logrus.Fatal(err)
	}

	if err = clientTo.Connect(ctx); err != nil {
		logrus.Fatalf("Can't connect to database, error: %s", err.Error())
	}

	dbFrom := clientFrom.Database("Schedule")
	dbTo := clientTo.Database("Studyum")

	collections := []string{"GeneralLessons", "Lessons", "Marks", "SignUpCodes", "StudyPlaces", "Users", "ParseJournalUsers", "ParseScheduleTypes"}
	for _, name := range collections {
		fmt.Println("Migrating " + name)
		if err = migrate(ctx, dbFrom, dbTo, name); err != nil {
			fmt.Println("Error " + err.Error())
		}
	}
}

func migrate(ctx context.Context, dbFrom, dbTo *mongo.Database, name string) error {
	collFrom := dbFrom.Collection(name)
	collTo := dbTo.Collection(name)

	cursor, err := collFrom.Find(ctx, bson.M{})
	if err != nil {
		return err
	}

	var records []interface{}
	if err = cursor.All(ctx, &records); err != nil {
		return err
	}

	if _, err = collTo.InsertMany(ctx, records); err != nil {
		return err
	}

	return nil
}
