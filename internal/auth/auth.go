package auth

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/auth/controllers"
	"studyum/internal/auth/handlers"
	"studyum/internal/auth/handlers/swagger"
	"studyum/internal/auth/repositories"
	codes "studyum/internal/codes/controllers"
	"studyum/pkg/encryption"
	controllers2 "studyum/pkg/jwt/controllers"
)

// @BasePath /api/user

//go:generate swag init --instanceName auth -o handlers/swagger -g auth.go -ot go,yaml
func New(core *gin.RouterGroup, codes codes.Controller, encryption encryption.Encryption, jwtController controllers2.Controller, db *mongo.Database) (handlers.Middleware, *handlers.Auth, *handlers.OAuth2) {
	swagger.SwaggerInfoauth.BasePath = "/api/user"

	usersCollection := db.Collection("Users")
	codesCollection := db.Collection("SignUpCodes")
	oauth2Collection := db.Collection("OAuth2Services")
	studyPlacesCollection := db.Collection("StudyPlaces")

	authRepository := repositories.NewAuth(usersCollection)
	codesRepository := repositories.NewCode(codesCollection)
	middlewareRepository := repositories.NewMiddleware(usersCollection, studyPlacesCollection)
	oauth2Repository := repositories.NewOAuth2(oauth2Collection, usersCollection)

	authController := controllers.NewAuth(jwtController, codes, encryption, authRepository, codesRepository)
	middlewareController := controllers.NewMiddleware(jwtController, middlewareRepository)
	oauth2Controller := controllers.NewOAuth2(oauth2Repository, encryption, jwtController)

	authMiddleware := handlers.NewMiddleware(middlewareController)
	authHandler := handlers.NewAuth(authMiddleware, authController, core)
	oauthHandler := handlers.NewOAuth2(authMiddleware, oauth2Controller, core.Group("/oauth2"))
	return authMiddleware, authHandler, oauthHandler
}
