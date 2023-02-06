package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"strconv"
	auth "studyum/internal/auth/handlers"
	"studyum/internal/general/controllers"
	"studyum/internal/general/handlers/swagger"
)

// @BasePath /api

//go:generate swag init --instanceName general -o swagger -g handler.go -ot go,yaml
type Handler interface {
	Uptime(ctx *gin.Context)
	Request(ctx *gin.Context)

	GetStudyPlaces(ctx *gin.Context)
	GetStudyPlaceByID(ctx *gin.Context)
	GetSelfStudyPlace(ctx *gin.Context)
}

type handler struct {
	auth.Middleware

	controller controllers.Controller
	Group      *gin.RouterGroup
}

func NewGeneralHandler(middleware auth.Middleware, controller controllers.Controller, group *gin.RouterGroup) Handler {
	h := &handler{Middleware: middleware, controller: controller, Group: group}

	group.GET("/studyPlaces", h.GetStudyPlaces)
	group.GET("/studyPlaces/:id", h.GetStudyPlaceByID)
	group.GET("/studyPlaces/self", h.MemberAuth(), h.GetStudyPlaceByID)

	group.GET("/uptime", h.Uptime)
	group.GET("/request", h.Request)

	swagger.SwaggerInfogeneral.BasePath = "/api"

	return h
}

// Uptime godoc
// @Router /uptime [get]
func (g *handler) Uptime(ctx *gin.Context) {
	ctx.JSON(200, "hi")
}

// Request godoc
// @Param host query string true "Host"
// @Router /request [get]
func (g *handler) Request(ctx *gin.Context) {
	response, err := http.Get("https://" + ctx.Query("host"))
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	_, _ = io.Copy(ctx.Writer, response.Body)
}

// GetStudyPlaces godoc
// @Router /studyPlaces [get]
func (g *handler) GetStudyPlaces(ctx *gin.Context) {
	isRestricted := ctx.Query("restricted")
	restricted, err := strconv.ParseBool(isRestricted)
	if err != nil {
		restricted = false
	}

	err, studyPlaces := g.controller.GetStudyPlaces(ctx, restricted)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, studyPlaces)
}

// GetStudyPlaceByID godoc
// @Param id path string true "Study Place ID"
// @Router /studyPlaces/{id} [get]
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
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, studyPlace)
}

// GetSelfStudyPlace godoc
// @Router /studyPlaces/self [get]
func (g *handler) GetSelfStudyPlace(ctx *gin.Context) {
	user := g.GetUser(ctx)

	err, studyPlace := g.controller.GetSelfStudyPlace(ctx, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, studyPlace)
}
