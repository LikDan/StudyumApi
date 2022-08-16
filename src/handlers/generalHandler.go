package handlers

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"studyum/src/controllers"
	"studyum/src/models"
)

type GeneralHandler struct {
	controller controllers.IGeneralController

	Group *gin.RouterGroup
}

func NewGeneralHandler(controller controllers.IGeneralController, group *gin.RouterGroup) *GeneralHandler {
	h := &GeneralHandler{controller: controller, Group: group}

	group.GET("/studyPlaces", h.GetStudyPlaces)
	group.GET("/uptime", h.Uptime)

	return h
}

func (g *GeneralHandler) Uptime(ctx *gin.Context) {
	ctx.JSON(200, "hi")
}

func (g *GeneralHandler) GetStudyPlaces(ctx *gin.Context) {
	err, studyPlaces := g.controller.GetStudyPlaces(ctx)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(http.StatusOK, studyPlaces)
}

func (g *GeneralHandler) RequestHandler(ctx *gin.Context) {
	response, err := http.Get("https://" + ctx.Query("host"))
	if models.BindError(err, 418, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	_, _ = io.Copy(ctx.Writer, response.Body)
}
