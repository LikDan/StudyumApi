package controllers

import (
	"github.com/gin-gonic/gin"
	"studyum/src/db"
	"studyum/src/models"
)

func GetStudyPlaces(ctx *gin.Context) {
	var studyPlaces []models.StudyPlace
	if err := db.GetStudyPlaces(&studyPlaces); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, studyPlaces)
}
