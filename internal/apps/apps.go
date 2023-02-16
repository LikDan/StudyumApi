package apps

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/apps/apps/kbp"
	"studyum/internal/apps/controllers"
	"studyum/internal/apps/entities"
	"studyum/internal/apps/repositories"
	"studyum/internal/apps/shared"
	"studyum/pkg/encryption"
)

func proceedApps() []entities.App {
	return []entities.App{
		&kbp.App{},
	}
}

func New(db *mongo.Database, encryption encryption.Encryption) controllers.Controller {
	apps := proceedApps()

	lessons := db.Collection("Lessons")
	users := db.Collection("Users")
	codeUsers := db.Collection("CodeUsers")

	for _, app := range apps {
		r := shared.NewShared(app.GetStudyPlaceID(context.Background()), encryption, lessons, users, codeUsers)
		app.Init(r)
	}

	appsRepository := repositories.NewApps(apps)
	dataRepository := repositories.NewData(db)
	controller := controllers.NewController(appsRepository, dataRepository)

	return controller
}
