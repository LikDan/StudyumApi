package controllers

import (
	"github.com/gin-gonic/gin"
	"studyum/src/db"
)

func GetStudyPlaces(ctx *gin.Context) {
	err, studyPlaces := db.GetStudyPlaces()
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, studyPlaces)
}
