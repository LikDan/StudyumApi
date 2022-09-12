package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"studyum/internal/controllers"
	"studyum/internal/controllers/validators"
	"studyum/internal/entities"
	"studyum/internal/handlers"
	pController "studyum/internal/parser/controller"
	pHandler "studyum/internal/parser/handler"
	pRepository "studyum/internal/parser/repository"
	"studyum/internal/repositories"
	fb "studyum/pkg/firebase"
	"studyum/pkg/jwt"
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

	firebaseCredentials := []byte(os.Getenv("FIREBASE_CREDENTIALS"))
	firebase := fb.NewFirebase(firebaseCredentials)

	parserRepository := pRepository.NewParserRepository(client)
	parserController := pController.NewParserController(parserRepository, firebase)
	parserHandler := pHandler.NewParserHandler(parserController)

	codesRepository := repositories.NewCodesRepository(client)
	signUpCodesRepository := repositories.NewSignUpCodesRepository(codesRepository)
	signUpCodesController := controllers.NewSignUpCodesController(signUpCodesRepository)

	secret := os.Getenv("JWT_SECRET")
	expTime := time.Minute * 10
	jwtController := jwt.New[entities.JWTClaims](expTime, secret)

	repository := repositories.NewRepository(client)
	userRepository := repositories.NewUserRepository(repository)
	generalRepository := repositories.NewGeneralRepository(repository)
	journalRepository := repositories.NewJournalRepository(repository)
	scheduleRepository := repositories.NewScheduleRepository(repository)

	scheduleValidator := validators.NewSchedule(validator.New())

	controller := controllers.NewController(jwtController, userRepository, generalRepository)
	userController := controllers.NewUserController(jwtController, signUpCodesController, userRepository)
	generalController := controllers.NewGeneralController(generalRepository)
	journalController := controllers.NewJournalController(parserHandler, journalRepository)
	scheduleController := controllers.NewScheduleController(parserHandler, scheduleValidator, scheduleRepository, generalController)

	jwtController.SetGetClaimsFunc(controller.GetClaims)

	engine := gin.Default()
	api := engine.Group("/api")

	handler := handlers.NewHandler(controller)
	handlers.NewGeneralHandler(handler, generalController, api)
	handlers.NewUserHandler(handler, userController, api.Group("/user"))
	handlers.NewJournalHandler(handler, journalController, api.Group("/journal"))
	handlers.NewScheduleHandler(handler, scheduleController, api.Group("/schedule"))

	logrus.Fatalf("Error launching server %s", engine.Run().Error())
}
