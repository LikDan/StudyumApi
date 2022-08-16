package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"studyum/src/controllers"
	"studyum/src/handlers"
	"studyum/src/parser"
	"studyum/src/parser/apps"
	"studyum/src/repositories"
	"studyum/src/utils"
	"time"
)

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL")))
	if err != nil {
		log.Fatal(err)
	}

	if err = client.Connect(nil); err != nil {
		log.Fatalf("Can't connect to database, error: %s", err.Error())
		return
	}

	repo := repositories.NewRepository(client)
	userRepository := repositories.NewUserRepository(repo)
	generalRepository := repositories.NewGeneralRepository(repo)
	journalRepository := repositories.NewJournalRepository(repo)
	scheduleRepository := repositories.NewScheduleRepository(repo)
	apps.Repository = repositories.NewParserRepository(repo)

	authController := controllers.NewAuthController(userRepository)
	userController := controllers.NewUserController(userRepository)
	generalController := controllers.NewGeneralController(generalRepository)
	journalController := controllers.NewJournalController(journalRepository)
	scheduleController := controllers.NewScheduleController(scheduleRepository)

	engine := gin.Default()
	api := engine.Group("/api")

	authHandler := handlers.NewAuthHandler(authController)
	handlers.NewGeneralHandler(generalController, api)
	handlers.NewUserHandler(authHandler, userController, api.Group("/user"))
	handlers.NewJournalHandler(authHandler, journalController, api.Group("/journal"))
	handlers.NewScheduleHandler(authHandler, scheduleController, api.Group("/schedule"))

	utils.InitFirebaseApp()
	parser.InitApps()

	if err = engine.Run(); err != nil {
		log.Fatalf("Error launching server %s", err.Error())
	}
}
