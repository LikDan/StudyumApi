package controllers

import (
	"github.com/gin-gonic/gin"
	"studyum/src/repositories"
)

type GeneralController struct {
	repository repositories.IGeneralRepository
}

func NewGeneralController(repository repositories.IGeneralRepository) *GeneralController {
	return &GeneralController{repository: repository}
}

func (g *GeneralController) GetStudyPlaces(ctx *gin.Context) {
	err, studyPlaces := g.repository.GetStudyPlaces(ctx)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, studyPlaces)
}
