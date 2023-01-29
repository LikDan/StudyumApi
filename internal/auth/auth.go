package auth

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/auth/controllers"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/handlers"
	"studyum/internal/auth/repositories"
	"studyum/internal/global"
	"studyum/pkg/encryption"
	"studyum/pkg/jwt"
)

func New(core *gin.RouterGroup, handler global.Handler, encryption encryption.Encryption, jwtController jwt.JWT[entities.JWTClaims], db *mongo.Database) (handlers.Middleware, *handlers.Auth, *handlers.OAuth2, controllers.Sessions) {
	usersCollection := db.Collection("Users")
	codesCollection := db.Collection("SignUpCodes")
	oauth2Collection := db.Collection("OAuth2Services")

	authRepository := repositories.NewAuth(usersCollection)
	codesRepository := repositories.NewCode(codesCollection)
	sessionsRepository := repositories.NewSessions(usersCollection)
	middlewareRepository := repositories.NewMiddleware(usersCollection)
	oauth2Repository := repositories.NewOAuth2(oauth2Collection, usersCollection)

	sessionsController := controllers.NewSessions(jwtController, sessionsRepository)
	authController := controllers.NewAuth(sessionsController, encryption, authRepository, codesRepository)
	middlewareController := controllers.NewMiddleware(jwtController, middlewareRepository)
	oauth2Controller := controllers.NewOAuth2(oauth2Repository, sessionsRepository, encryption, jwtController)

	authMiddleware := handlers.NewMiddleware(handler, middlewareController)
	authHandler := handlers.NewAuth(handler, authMiddleware, authController, core)
	oauthHandler := handlers.NewOAuth2(handler, authMiddleware, oauth2Controller, core.Group("/oauth2"))
	return authMiddleware, authHandler, oauthHandler, sessionsController
}
