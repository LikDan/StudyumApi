package user

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	auth "studyum/internal/auth/handlers"
	codes "studyum/internal/codes/controllers"
	"studyum/internal/user/controllers"
	"studyum/internal/user/handlers"
	"studyum/internal/user/handlers/swagger"
	"studyum/internal/user/repositories"
	"studyum/pkg/encryption"
	jwt "studyum/pkg/jwt/controllers"
)

// @BasePath /api/user

//go:generate swag init --instanceName user -o handlers/swagger -g user.go -ot go,yaml
func New(core *gin.RouterGroup, auth auth.Middleware, encrypt encryption.Encryption, codesController codes.Controller, sessionsController jwt.Controller, db *mongo.Database) handlers.Handler {
	swagger.SwaggerInfouser.BasePath = "/api/user"

	users := db.Collection("Users")
	signUpCodes := db.Collection("SignUpCodes")

	repository := repositories.NewUserRepository(users, signUpCodes)

	controller := controllers.NewUserController(repository, codesController, sessionsController, encrypt)

	handler := handlers.NewUserHandler(auth, controller, core)
	return handler
}
