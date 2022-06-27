package parser

import (
	"studyum/src/api"
	"studyum/src/models"
)

func GetAppByStudyPlaceId(id int, app *models.IParserApp) *models.Error {
	for _, app_ := range Apps {
		if app_.GetStudyPlaceId() == id {
			*app = app_
			break
		}
	}

	if app == nil {
		models.BindErrorStr("not authorized", 401, api.UNDEFINED)
	}

	return models.EmptyError()
}
