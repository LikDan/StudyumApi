package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"studyum/internal/apps"
	"studyum/internal/journal/entities"
)

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL")))
	if err != nil {
		logrus.Fatal(err)
	}

	ctx := context.Background()
	if err = client.Connect(ctx); err != nil {
		logrus.Fatalf("Can't connect to database, error: %s", err.Error())
	}

	db := client.Database("Studyum")

	c := apps.New(db)
	id, _ := primitive.ObjectIDFromHex("631261e11b8b855cc75cec35")
	id2, _ := primitive.ObjectIDFromHex("63e37313dbbdba2d74ad3ea2")
	c.Event(id, "AddMark", entities.Lesson{Id: id2})
	fmt.Println("program")
}
