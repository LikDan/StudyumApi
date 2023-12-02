package schedule

import (
	"github.com/gin-gonic/gin"
	v "github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
	apps "studyum/internal/apps/controllers"
	auth "studyum/internal/auth/handlers"
	general "studyum/internal/general/controllers"
	"studyum/internal/schedule/controllers"
	"studyum/internal/schedule/controllers/validators"
	"studyum/internal/schedule/handlers"
	"studyum/internal/schedule/handlers/swagger"
	"studyum/internal/schedule/repositories"
)

// @BasePath /api/schedule

//go:generate swag init --instanceName schedule -o handlers/swagger -g schedule.go -ot go,yaml
func New(core *gin.RouterGroup, auth auth.Middleware, apps apps.Controller, general general.Controller, db *mongo.Database) handlers.Handler {
	swagger.SwaggerInfoschedule.BasePath = "/api/schedule"

	studyPlaces := db.Collection("StudyPlaces")
	lessons := db.Collection("Lessons")
	generalLessons := db.Collection("GeneralLessons")
	schedule := db.Collection("Schedule")

	repository := repositories.NewScheduleRepository(studyPlaces, lessons, generalLessons, schedule, db)
	generalLessonsRepository := repositories.NewGeneralLessonsRepository(generalLessons)

	validator := validators.NewSchedule(v.New())
	controller := controllers.NewController(repository, generalLessonsRepository, general, apps, validator)

	handler := handlers.NewScheduleHandler(auth, controller, core)
	return handler
}
