package auth

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"studyum/internal/auth/controllers"
	"studyum/internal/auth/handlers"
	"studyum/internal/auth/handlers/swagger"
	"studyum/internal/auth/repositories"
	codes "studyum/internal/codes/controllers"
	"studyum/internal/utils/jwt"
	"studyum/pkg/encryption"
)

// @BasePath /api/user

//go:generate swag init --instanceName auth -o handlers/swagger -g auth.go -ot go,yaml
func New(core *gin.RouterGroup, grpcServer *grpc.Server, codes codes.Controller, encryption encryption.Encryption, jwtController jwt.JWT, db *mongo.Database) (handlers.Middleware, *handlers.Auth, *handlers.OAuth2) {
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
	authHandler := handlers.NewAuth(authMiddleware, authController, core, grpcServer)
	oauthHandler := handlers.NewOAuth2(authMiddleware, oauth2Controller, core.Group("/oauth2"))
	return authMiddleware, authHandler, oauthHandler
}
