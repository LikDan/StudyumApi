package controllers

import (
	"github.com/pkg/errors"
	apps "studyum/internal/apps/controllers"
	"studyum/internal/general/controllers"
	"studyum/internal/schedule/controllers/validators"
	"studyum/internal/schedule/repositories"
)

var NotValidParams = errors.New("not valid params")
var NoPermission = errors.New("no permission")

type Controller struct {
	ScheduleController
	GeneralLessons GeneralLessonsController
}

func NewController(repository repositories.Repository, generalLessonsRepository repositories.GeneralLessonsRepository, studyPlacesController controllers.Controller, apps apps.Controller, validator validators.Validator) Controller {
	return Controller{
		ScheduleController: NewScheduleController(repository, studyPlacesController, apps, validator),
		GeneralLessons:     NewGeneralLessonController(generalLessonsRepository, studyPlacesController, apps, validator),
	}
}
