package apps

import (
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/apps/apps/kbp"
	"studyum/internal/apps/controllers"
	"studyum/internal/apps/entities"
	"studyum/internal/apps/repositories"
)

func proceedApps() []entities.App {
	return []entities.App{
		kbp.NewApp(),
	}
}

func New(db *mongo.Database) controllers.Controller {
	apps := proceedApps()

	appsRepository := repositories.NewApps(apps)
	dataRepository := repositories.NewData(db)
	controller := controllers.NewController(appsRepository, dataRepository)

	return controller
}
