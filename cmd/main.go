package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"studyum/internal/controllers"
	"studyum/internal/handlers"
	pController "studyum/internal/parser/controller"
	pHandler "studyum/internal/parser/handler"
	pRepository "studyum/internal/parser/repository"
	"studyum/internal/repositories"
	fb "studyum/pkg/firebase"
	"time"
)

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL")))
	if err != nil {
		logrus.Fatal(err)
	}

	ctx := context.Background()
	if err = client.Connect(ctx); err != nil {
		logrus.Fatalf("Can't connect to database, error: %s", err.Error())
	}

	repository := repositories.NewRepository(client)
	userRepository := repositories.NewUserRepository(repository)
	generalRepository := repositories.NewGeneralRepository(repository)
	journalRepository := repositories.NewJournalRepository(repository)
	scheduleRepository := repositories.NewScheduleRepository(repository)

	parserRepository := pRepository.NewParserRepository(client)

	controller := controllers.NewController(userRepository)
	userController := controllers.NewUserController(userRepository)
	generalController := controllers.NewGeneralController(generalRepository)
	journalController := controllers.NewJournalController(journalRepository)
	scheduleController := controllers.NewScheduleController(scheduleRepository)

	parserController := pController.NewParserController(parserRepository)

	firebaseCredentials := []byte(os.Getenv("FIREBASE_CREDENTIALS"))
	firebase := fb.NewFirebase(firebaseCredentials)

	engine := gin.Default()
	api := engine.Group("/api")

	handler := handlers.NewHandler(controller)
	handlers.NewGeneralHandler(handler, generalController, api)
	handlers.NewUserHandler(handler, userController, api.Group("/user"))
	handlers.NewJournalHandler(handler, journalController, api.Group("/journal"))
	handlers.NewScheduleHandler(handler, scheduleController, api.Group("/schedule"))

	pHandler.NewParserHandler(firebase, parserController)

	logrus.Fatalf("Error launching server %s", engine.Run().Error())
}
