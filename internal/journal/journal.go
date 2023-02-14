package journal

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	apps "studyum/internal/apps/controllers"
	auth "studyum/internal/auth/handlers"
	"studyum/internal/journal/controllers"
	"studyum/internal/journal/handlers"
	"studyum/internal/journal/handlers/swagger"
	"studyum/internal/journal/repositories"
	"studyum/pkg/encryption"
)

// @BasePath /api/journal

//go:generate swag init --instanceName journal -o handlers/swagger -g journal.go -ot go,yaml
func New(core *gin.RouterGroup, auth auth.Middleware, apps apps.Controller, encrypt encryption.Encryption, db *mongo.Database) handlers.Handler {
	swagger.SwaggerInfojournal.BasePath = "/api/journal"

	users := db.Collection("Users")
	lessons := db.Collection("Lessons")
	studyPlaces := db.Collection("StudyPlaces")

	repository := repositories.NewJournalRepository(users, lessons, studyPlaces)

	queryController := controllers.NewJournalController(repository, encrypt)
	controller := controllers.NewController(queryController, repository, encrypt, apps)

	handler := handlers.NewJournalHandler(auth, controller, queryController, core)
	return handler
}
