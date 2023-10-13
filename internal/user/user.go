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
	"studyum/internal/utils/jwt"
	"studyum/pkg/encryption"
)

// @BasePath /api/user

//go:generate swag init --instanceName user -o handlers/swagger -g user.go -ot go,yaml
func New(core *gin.RouterGroup, auth auth.Middleware, encrypt encryption.Encryption, codesController codes.Controller, sessionsController jwt.JWT, db *mongo.Database) (handlers.Handler, controllers.Controller) {
	swagger.SwaggerInfouser.BasePath = "/api/user"

	users := db.Collection("Users")
	signUpCodes := db.Collection("SignUpCodes")
	preferences := db.Collection("UserPreferences")

	repository := repositories.NewUserRepository(users, signUpCodes)
	preferencesRepository := repositories.NewPreferencesRepository(preferences)

	controller := controllers.NewUserController(repository, codesController, sessionsController, encrypt)
	preferencesController := controllers.NewPreferencesController(preferencesRepository)

	handler := handlers.NewUserHandler(auth, controller, preferencesController, core)
	return handler, controller
}
