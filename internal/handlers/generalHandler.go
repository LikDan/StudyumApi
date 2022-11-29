package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"strconv"
	"studyum/internal/controllers"
)

type GeneralHandler interface {
	Uptime(ctx *gin.Context)

	GetStudyPlaces(ctx *gin.Context)
	GetStudyPlaceByID(ctx *gin.Context)

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
	group.GET("/studyPlaces/:id", h.GetStudyPlaceByID)

	group.GET("/uptime", h.Uptime)
	group.GET("/request", h.Request)

	return h
}

func (g *generalHandler) Uptime(ctx *gin.Context) {
	ctx.JSON(200, "hi")
}

func (g *generalHandler) GetStudyPlaces(ctx *gin.Context) {
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

func (g *generalHandler) GetStudyPlaceByID(ctx *gin.Context) {
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

func (g *generalHandler) Request(ctx *gin.Context) {
	response, err := http.Get("https://" + ctx.Query("host"))
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	_, _ = io.Copy(ctx.Writer, response.Body)
}
