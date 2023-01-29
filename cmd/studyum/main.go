package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"studyum/internal/auth"
	authEntries "studyum/internal/auth/entities"
	controllers2 "studyum/internal/general/controllers"
	handlers2 "studyum/internal/general/handlers"
	repositories2 "studyum/internal/general/repositories"
	"studyum/internal/global"
	"studyum/internal/journal/controllers"
	"studyum/internal/journal/handlers"
	"studyum/internal/journal/repositories"
	pController "studyum/internal/parser/controller"
	pHandler "studyum/internal/parser/handler"
	pRepository "studyum/internal/parser/repository"
	"studyum/internal/schedule"
	"studyum/internal/user"
	"studyum/pkg/encryption"
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

	defer logrus.Warning("Studyum is stopping at", time.Now().Format("2006-01-02 15:04"))

	firebaseCredentials := []byte(os.Getenv("FIREBASE_CREDENTIALS"))
	firebase := fb.NewFirebase(firebaseCredentials)
	encrypt := encryption.NewEncryption(os.Getenv("ENCRYPTION_SECRET"))

	parserRepository := pRepository.NewParserRepository(client)
	parserController := pController.NewParserController(parserRepository, encrypt, firebase)
	parserHandler := pHandler.NewParserHandler(parserController)

	secret := os.Getenv("JWT_SECRET")
	expTime := time.Minute * 10
	jwtController := jwt.New[authEntries.JWTClaims](expTime, secret)

	handler := global.NewHandler()

	engine := gin.Default()
	api := engine.Group("/api")

	db := client.Database("Studyum")

	userRepository := user.NewUserRepository(db.Collection("Users"), db.Collection("SignUpCodes"))
	generalRepository := repositories2.NewGeneralRepository(db.Collection("StudyPlaces"))
	journalRepository := repositories.NewJournalRepository(db.Collection("Users"), db.Collection("Lessons"), db.Collection("StudyPlaces"))
	scheduleRepository := schedule.NewScheduleRepository(db.Collection("StudyPlaces"), db.Collection("Lessons"), db.Collection("GeneralLessons"))

	scheduleValidator := schedule.NewSchedule(validator.New())

	authMiddleware, _, _, sessionsController := auth.New(api.Group("/user"), handler, encrypt, jwtController, db)

	userController := user.NewUserController(userRepository, sessionsController, encrypt, parserHandler)
	generalController := controllers2.NewGeneralController(generalRepository)
	journalController := controllers.NewJournalController(journalRepository, encrypt)
	mainJournalController := controllers.NewController(parserHandler, journalController, journalRepository, encrypt)
	scheduleController := schedule.NewScheduleController(parserHandler, scheduleValidator, scheduleRepository, generalController)

	handlers2.NewGeneralHandler(handler, generalController, api)
	user.NewUserHandler(handler, authMiddleware, userController, api.Group("/user"))
	handlers.NewJournalHandler(handler, authMiddleware, mainJournalController, journalController, api.Group("/journal"))
	schedule.NewScheduleHandler(handler, authMiddleware, scheduleController, api.Group("/schedule"))

	logrus.Fatalf("Error launching server %s", engine.Run().Error())
}
