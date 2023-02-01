package auth

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/auth/controllers"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/handlers"
	"studyum/internal/auth/repositories"
	"studyum/pkg/encryption"
	"studyum/pkg/jwt"
	"studyum/pkg/mail"
)

func New(core *gin.RouterGroup, encryption encryption.Encryption, jwtController jwt.JWT[entities.JWTClaims], mail mail.Mail, db *mongo.Database) (handlers.Middleware, *handlers.Auth, *handlers.OAuth2, controllers.Sessions) {
	usersCollection := db.Collection("Users")
	codesCollection := db.Collection("SignUpCodes")
	verificationCodesCollection := db.Collection("VerificationCodes")
	oauth2Collection := db.Collection("OAuth2Services")
	studyPlacesCollection := db.Collection("StudyPlaces")

	authRepository := repositories.NewAuth(usersCollection)
	codesRepository := repositories.NewCode(codesCollection)
	verificationCodesRepository := repositories.NewVerificationCodes(verificationCodesCollection)
	sessionsRepository := repositories.NewSessions(usersCollection)
	middlewareRepository := repositories.NewMiddleware(usersCollection, studyPlacesCollection)
	oauth2Repository := repositories.NewOAuth2(oauth2Collection, usersCollection)

	sessionsController := controllers.NewSessions(jwtController, sessionsRepository)
	authController := controllers.NewAuth(sessionsController, mail, encryption, authRepository, codesRepository, verificationCodesRepository)
	middlewareController := controllers.NewMiddleware(jwtController, middlewareRepository)
	oauth2Controller := controllers.NewOAuth2(oauth2Repository, sessionsRepository, encryption, jwtController)

	authMiddleware := handlers.NewMiddleware(middlewareController)
	authHandler := handlers.NewAuth(authMiddleware, authController, core)
	oauthHandler := handlers.NewOAuth2(authMiddleware, oauth2Controller, core.Group("/oauth2"))
	return authMiddleware, authHandler, oauthHandler, sessionsController
}
