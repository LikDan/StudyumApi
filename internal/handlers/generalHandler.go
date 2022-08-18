package handlers

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"studyum/internal/controllers"
)

type GeneralHandler interface {
	Uptime(ctx *gin.Context)
	GetStudyPlaces(ctx *gin.Context)
	Request(ctx *gin.Context)
}

type generalHandler struct {
	Handler

	controller controllers.GeneralController

	Group *gin.RouterGroup
}

func NewGeneralHandler(handler Handler, controller controllers.GeneralController, group *gin.RouterGroup) GeneralHandler {
	h := &generalHandler{Handler: handler, controller: controller, Group: group}

	group.GET("/studyPlaces", h.GetStudyPlaces)
	group.GET("/uptime", h.Uptime)
	group.GET("/request", h.Request)

	return h
}

func (g *generalHandler) Uptime(ctx *gin.Context) {
	ctx.JSON(200, "hi")
}

func (g *generalHandler) GetStudyPlaces(ctx *gin.Context) {
	err, studyPlaces := g.controller.GetStudyPlaces(ctx)
	if err != nil {
		g.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, studyPlaces)
}

func (g *generalHandler) Request(ctx *gin.Context) {
	response, err := http.Get("https://" + ctx.Query("host"))
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	_, _ = io.Copy(ctx.Writer, response.Body)
}
