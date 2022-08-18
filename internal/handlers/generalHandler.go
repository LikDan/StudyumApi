package handlers

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"studyum/internal/controllers"
)

type GeneralHandler struct {
	IHandler

	controller controllers.GeneralController

	Group *gin.RouterGroup
}

func NewGeneralHandler(handler IHandler, controller controllers.GeneralController, group *gin.RouterGroup) *GeneralHandler {
	h := &GeneralHandler{IHandler: handler, controller: controller, Group: group}

	group.GET("/studyPlaces", h.GetStudyPlaces)
	group.GET("/uptime", h.Uptime)

	return h
}

func (g *GeneralHandler) Uptime(ctx *gin.Context) {
	ctx.JSON(200, "hi")
}

func (g *GeneralHandler) GetStudyPlaces(ctx *gin.Context) {
	err, studyPlaces := g.controller.GetStudyPlaces(ctx)
	if err != nil {
		g.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, studyPlaces)
}

func (g *GeneralHandler) RequestHandler(ctx *gin.Context) {
	response, err := http.Get("https://" + ctx.Query("host"))
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	_, _ = io.Copy(ctx.Writer, response.Body)
}
