package codes

import (
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/codes/controllers"
	"studyum/internal/codes/repositories"
	"studyum/pkg/mail"
	"time"
)

func New(expireTime time.Duration, timeout time.Duration, mailer mail.Mail, db *mongo.Database) controllers.Controller {
	collection := db.Collection("VerificationCodes")

	repository := repositories.New(collection)
	controller := controllers.New(repository, mailer, expireTime, timeout)

	return controller
}
