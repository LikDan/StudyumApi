package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"strconv"
	"studyum/internal/general/controllers"
	"studyum/internal/global"
)

type Handler interface {
	Uptime(ctx *gin.Context)

	GetStudyPlaces(ctx *gin.Context)
	GetStudyPlaceByID(ctx *gin.Context)

	Request(ctx *gin.Context)
}

type handler struct {
	global.Handler

	controller controllers.Controller

	Group *gin.RouterGroup
}

func NewGeneralHandler(globalHandler global.Handler, controller controllers.Controller, group *gin.RouterGroup) Handler {
	h := &handler{Handler: globalHandler, controller: controller, Group: group}

	group.GET("/studyPlaces", h.GetStudyPlaces)
	group.GET("/studyPlaces/:id", h.GetStudyPlaceByID)

	group.GET("/uptime", h.Uptime)
	group.GET("/request", h.Request)

	return h
}

func (g *handler) Uptime(ctx *gin.Context) {
	ctx.JSON(200, "hi")
}

func (g *handler) GetStudyPlaces(ctx *gin.Context) {
	isRestricted := ctx.Query("restricted")
	restricted, err := strconv.ParseBool(isRestricted)
	if err != nil {
		restricted = false
	}

	err, studyPlaces := g.controller.GetStudyPlaces(ctx, restricted)
	if err != nil {
		g.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, studyPlaces)
}

func (g *handler) GetStudyPlaceByID(ctx *gin.Context) {
	isRestricted := ctx.Query("restricted")
	restricted, err := strconv.ParseBool(isRestricted)
	if err != nil {
		restricted = false
	}

	idHex := ctx.Param("id")
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err, studyPlace := g.controller.GetStudyPlaceByID(ctx, id, restricted)
	if err != nil {
		g.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, studyPlace)
}

func (g *handler) Request(ctx *gin.Context) {
	response, err := http.Get("https://" + ctx.Query("host"))
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	_, _ = io.Copy(ctx.Writer, response.Body)
}
