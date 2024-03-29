package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"net/http"
	"strconv"
	"studyum/grpc/studyPlaces/protostudyplaces"
	auth "studyum/internal/auth/handlers"
	"studyum/internal/general/controllers"
	"studyum/internal/general/handlers/swagger"
)

// @BasePath /api

//go:generate swag init --instanceName general -o swagger -g handler.go -ot go,yaml
type Handler interface {
	GetStudyPlaces(ctx *gin.Context)
	GetStudyPlaceByID(ctx *gin.Context)
	GetSelfStudyPlace(ctx *gin.Context)
}

type handler struct {
	auth.Middleware

	controller controllers.Controller
	Group      *gin.RouterGroup
}

func (g *handler) GetByID(ctx context.Context, request *protostudyplaces.IdRequest) (*protostudyplaces.StudyPlace, error) {
	id, err := primitive.ObjectIDFromHex(request.Id)
	if err != nil {
		return nil, err
	}

	err, studyPlace := g.controller.GetStudyPlaceByID(ctx, id, false)
	if err != nil {
		return nil, err
	}

	return &protostudyplaces.StudyPlace{
		Name:       studyPlace.Name,
		Restricted: studyPlace.Restricted,
	}, nil
}

func NewGeneralHandler(middleware auth.Middleware, controller controllers.Controller, group *gin.RouterGroup, grpcServer *grpc.Server) Handler {
	h := &handler{Middleware: middleware, controller: controller, Group: group}

	protostudyplaces.RegisterStudyPlacesServer(grpcServer, h)

	group.GET("/studyPlaces", h.GetStudyPlaces)
	group.GET("/studyPlaces/:id", h.GetStudyPlaceByID)
	group.GET("/studyPlaces/self", h.MemberAuth(), h.GetSelfStudyPlace)

	swagger.SwaggerInfogeneral.BasePath = "/api"

	return h
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
