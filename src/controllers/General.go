package controllers

import (
	"github.com/gin-gonic/gin"
	"studyum/src/repositories"
)

var GeneralRepository repositories.IGeneralRepository

func GetStudyPlaces(ctx *gin.Context) {
	err, studyPlaces := GeneralRepository.GetStudyPlaces(ctx)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, studyPlaces)
}
