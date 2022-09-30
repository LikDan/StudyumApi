package main

import (
	"context"
	"fmt"
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
	"studyum/pkg/encryption"
	fb "studyum/pkg/firebase"
	"studyum/pkg/jwt"
	"time"
)

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	a := 1209600000000000
	fmt.Println(a)

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

	secret := os.Getenv("JWT_SECRET")
	expTime := time.Minute * 10
	jwtController := jwt.New[entities.JWTClaims](expTime, secret)

	repository := repositories.NewRepository(client)
	userRepository := repositories.NewUserRepository(repository)
	generalRepository := repositories.NewGeneralRepository(repository)
	journalRepository := repositories.NewJournalRepository(repository)
	scheduleRepository := repositories.NewScheduleRepository(repository)
	signUpCodesRepository := repositories.NewSignUpCodesRepository(repository)

	scheduleValidator := validators.NewSchedule(validator.New())
	encrypt := encryption.NewEncryption(os.Getenv("ENCRYPTION_SECRET"))

	signUpCodesController := controllers.NewSignUpCodesController(signUpCodesRepository)
	controller := controllers.NewController(jwtController, userRepository, generalRepository, encrypt)
	userController := controllers.NewUserController(jwtController, signUpCodesController, userRepository, encrypt, parserHandler)
	generalController := controllers.NewGeneralController(generalRepository)
	journalController := controllers.NewJournalController(parserHandler, journalRepository, encrypt)
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
