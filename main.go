package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"studyum/src/controllers"
	"studyum/src/handlers"
	pController "studyum/src/parser/controller"
	pHandler "studyum/src/parser/handler"
	"studyum/src/parser/repository"
	"studyum/src/repositories"
	"studyum/src/utils"
	"time"
)

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL")))
	if err != nil {
		logrus.Fatal(err)
	}

	if err = client.Connect(nil); err != nil {
		logrus.Fatalf("Can't connect to database, error: %s", err.Error())
		return
	}

	repo := repositories.NewRepository(client)
	userRepository := repositories.NewUserRepository(repo)
	generalRepository := repositories.NewGeneralRepository(repo)
	journalRepository := repositories.NewJournalRepository(repo)
	scheduleRepository := repositories.NewScheduleRepository(repo)
	parserRepository := repository.NewParserRepository(client)

	controller := controllers.NewController(userRepository)
	userController := controllers.NewUserController(userRepository)
	generalController := controllers.NewGeneralController(generalRepository)
	journalController := controllers.NewJournalController(journalRepository)
	scheduleController := controllers.NewScheduleController(scheduleRepository)
	parserController := pController.NewParserController(parserRepository)

	engine := gin.Default()
	api := engine.Group("/api")

	handler := handlers.NewHandler(controller)
	handlers.NewGeneralHandler(handler, generalController, api)
	handlers.NewUserHandler(handler, userController, api.Group("/user"))
	handlers.NewJournalHandler(handler, journalController, api.Group("/journal"))
	handlers.NewScheduleHandler(handler, scheduleController, api.Group("/schedule"))

	pHandler.NewParserHandler(parserController)

	utils.InitFirebaseApp()

	if err = engine.Run(); err != nil {
		logrus.Fatalf("Error launching server %s", err.Error())
	}
}
