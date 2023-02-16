package kbp

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"studyum/internal/apps/apps/kbp/marks"
	"studyum/internal/apps/apps/kbp/schedule"
	"studyum/internal/apps/apps/kbp/shared"
	"studyum/internal/apps/entities"
	appShared "studyum/internal/apps/shared"
)

type App struct {
	entities.LessonsManageInterface
	entities.MarksManageInterface
}

func (a *App) Init(repository appShared.Shared) {
	db, err := a.initDB()
	if err != nil {
		logrus.Error("db initialization error: " + err.Error())
		return
	}

	login := os.Getenv("KBP_LOGIN")
	password := os.Getenv("KBP_PASSWORD")
	auth := shared.NewAuthRepository(login, password)

	a.LessonsManageInterface = schedule.New(repository, db, auth)
	a.MarksManageInterface = marks.New(repository, auth)
}

func (a *App) GetStudyPlaceID(context.Context) primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex("631261e11b8b855cc75cec35")
	return id
}

func (a *App) initDB() (*mongo.Database, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL")))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if err = client.Connect(ctx); err != nil {
		return nil, err
	}

	return client.Database("Kbp"), nil
}
