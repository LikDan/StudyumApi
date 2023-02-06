package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"studyum/internal/auth"
	authEntries "studyum/internal/auth/entities"
	"studyum/internal/codes"
	"studyum/internal/general"
	"studyum/internal/journal"
	pController "studyum/internal/parser/controller"
	pHandler "studyum/internal/parser/handler"
	pRepository "studyum/internal/parser/repository"
	"studyum/internal/schedule"
	"studyum/internal/user"
	"studyum/internal/utils/middlewares"
	"studyum/pkg/encryption"
	fb "studyum/pkg/firebase"
	"studyum/pkg/jwt"
	"studyum/pkg/mail"
	"time"
)

//go:generate go generate studyum/internal/auth
func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	if gin.Mode() == gin.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL")))
	if err != nil {
		logrus.Fatal(err)
	}

	ctx := context.Background()
	if err = client.Connect(ctx); err != nil {
		logrus.Fatalf("Can't connect to database, error: %s", err.Error())
	}

	id := os.Getenv("GMAIL_CLIENT_ID")
	secret := os.Getenv("GMAIL_CLIENT_SECRET")
	access := os.Getenv("GMAIL_ACCESS_TOKEN")
	refresh := os.Getenv("GMAIL_REFRESH_TOKEN")
	mailer := mail.NewMail(context.Background(), mail.Mode(gin.Mode()), id, secret, access, refresh, "email-templates")

	mailer.ForceSend("likdan.official@gmail.com", "Application started", "Studyum app has been started")
	defer mailer.ForceSend("likdan.official@gmail.com", "Application stopped", "Studyum app has been stopped at"+time.Now().Format("2006-01-02 15:04"))

	defer logrus.Warning("Studyum is stopping at", time.Now().Format("2006-01-02 15:04"))

	firebaseCredentials := []byte(os.Getenv("FIREBASE_CREDENTIALS"))
	firebase := fb.NewFirebase(firebaseCredentials)
	encrypt := encryption.NewEncryption(os.Getenv("ENCRYPTION_SECRET"))

	parserRepository := pRepository.NewParserRepository(client)
	parserController := pController.NewParserController(parserRepository, encrypt, firebase)
	parserHandler := pHandler.NewParserHandler(parserController)

	secret = os.Getenv("JWT_SECRET")
	expTime := time.Minute * 10
	jwtController := jwt.New[authEntries.JWTClaims](expTime, secret)

	engine := gin.Default()
	engine.Use(middlewares.ErrorMiddleware())
	api := engine.Group("/api")

	db := client.Database("Studyum")

	codesController := codes.New(time.Minute*15, time.Minute, mailer, db)
	authMiddleware, _, _, sessionsController := auth.New(api.Group("/user"), codesController, encrypt, jwtController, db)

	_, generalController := general.New(api, authMiddleware, db)
	_ = journal.New(api.Group("/journal"), authMiddleware, parserHandler, encrypt, db)
	_ = schedule.New(api.Group("/schedule"), authMiddleware, parserHandler, generalController, db)
	_ = user.New(api.Group("/user"), authMiddleware, encrypt, parserHandler, codesController, sessionsController, db)

	loadSwagger(engine, "general", "auth", "user", "schedule", "journal")

	logrus.Fatalf("Error launching server %s", engine.Run().Error())
}

func loadSwagger(e *gin.Engine, names ...string) {
	for _, name := range names {
		s := ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.InstanceName(name))
		e.GET("/swagger/"+name+"/*any", s)
	}
}
