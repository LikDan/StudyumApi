package general

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	auth "studyum/internal/auth/handlers"
	"studyum/internal/general/controllers"
	"studyum/internal/general/handlers"
	"studyum/internal/general/repositories"
)

func New(core *gin.RouterGroup, grpcServer *grpc.Server, auth auth.Middleware, db *mongo.Database) (handlers.Handler, controllers.Controller) {
	studyPlaces := db.Collection("StudyPlaces")

	repository := repositories.NewGeneralRepository(studyPlaces)
	controller := controllers.NewGeneralController(repository)
	handler := handlers.NewGeneralHandler(auth, controller, core, grpcServer)

	return handler, controller
}
