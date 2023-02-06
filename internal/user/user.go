package user

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	authControllers "studyum/internal/auth/controllers"
	auth "studyum/internal/auth/handlers"
	codes "studyum/internal/codes/controllers"
	parser "studyum/internal/parser/handler"
	"studyum/internal/user/controllers"
	"studyum/internal/user/handlers"
	"studyum/internal/user/handlers/swagger"
	"studyum/internal/user/repositories"
	"studyum/pkg/encryption"
)

// @BasePath /api/schedule

//go:generate swag init --instanceName user -o handlers/swagger -g user.go -ot go,yaml
func New(core *gin.RouterGroup, auth auth.Middleware, encrypt encryption.Encryption, apps parser.Handler, codesController codes.Controller, sessionsController authControllers.Sessions, db *mongo.Database) handlers.Handler {
	swagger.SwaggerInfouser.BasePath = "/api/schedule"

	users := db.Collection("Users")
	signUpCodes := db.Collection("SignUpCodes")

	repository := repositories.NewUserRepository(users, signUpCodes)

	controller := controllers.NewUserController(repository, codesController, sessionsController, encrypt, apps)

	handler := handlers.NewUserHandler(auth, controller, core)
	return handler
}
