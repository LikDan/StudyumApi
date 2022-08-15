package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"net/http"
	"os"
	"studyum/src/controllers"
	"studyum/src/models"
	"studyum/src/parser"
	"studyum/src/parser/apps"
	"studyum/src/repositories"
	"studyum/src/routes"
	"studyum/src/utils"
	"time"
)

func uptimeHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "hi"})
}

func requestHandler(ctx *gin.Context) {
	response, err := http.Get("https://" + ctx.Query("host"))
	if models.BindError(err, 418, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	_, _ = io.Copy(ctx.Writer, response.Body)
}

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
	controllers.GeneralRepository = repositories.NewGeneralRepository(repo)
	controllers.JournalRepository = repositories.NewJournalRepository(repo)
	controllers.ScheduleRepository = repositories.NewScheduleRepository(repo)
	controllers.UserRepository = repositories.NewUserRepository(repo)
	apps.Repository = repositories.NewParserRepository(repo)

	utils.InitFirebaseApp()
	parser.InitApps()

	r := gin.Default()

	r.HEAD("/api", uptimeHandler)
	r.GET("/request", requestHandler)

	routes.Bind(r)

	log.Info("Application launched")

	if err = r.Run(); err != nil {
		log.Fatalf("Error launching server %s", err.Error())
	}
}
