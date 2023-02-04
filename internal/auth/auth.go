package auth

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/auth/controllers"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/handlers"
	"studyum/internal/auth/repositories"
	codes "studyum/internal/codes/controllers"
	"studyum/pkg/encryption"
	"studyum/pkg/jwt"
)

func New(core *gin.RouterGroup, codes codes.Controller, encryption encryption.Encryption, jwtController jwt.JWT[entities.JWTClaims], db *mongo.Database) (handlers.Middleware, *handlers.Auth, *handlers.OAuth2, controllers.Sessions) {
	usersCollection := db.Collection("Users")
	codesCollection := db.Collection("SignUpCodes")
	oauth2Collection := db.Collection("OAuth2Services")
	studyPlacesCollection := db.Collection("StudyPlaces")

	authRepository := repositories.NewAuth(usersCollection)
	codesRepository := repositories.NewCode(codesCollection)
	sessionsRepository := repositories.NewSessions(usersCollection)
	middlewareRepository := repositories.NewMiddleware(usersCollection, studyPlacesCollection)
	oauth2Repository := repositories.NewOAuth2(oauth2Collection, usersCollection)

	sessionsController := controllers.NewSessions(jwtController, sessionsRepository)
	authController := controllers.NewAuth(sessionsController, codes, encryption, authRepository, codesRepository)
	middlewareController := controllers.NewMiddleware(jwtController, middlewareRepository)
	oauth2Controller := controllers.NewOAuth2(oauth2Repository, sessionsRepository, encryption, jwtController)

	authMiddleware := handlers.NewMiddleware(middlewareController)
	authHandler := handlers.NewAuth(authMiddleware, authController, core)
	oauthHandler := handlers.NewOAuth2(authMiddleware, oauth2Controller, core.Group("/oauth2"))
	return authMiddleware, authHandler, oauthHandler, sessionsController
}
