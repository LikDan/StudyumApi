package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/internal/general/handlers/swagger"
	"studyum/internal/i18n/controllers"
)

type PublicHandler interface {
	GetDefault(ctx *gin.Context)
}

type handler struct {
	controller controllers.Controller
	Group      *gin.RouterGroup
}

func NewPublicHandler(controller controllers.Controller, group *gin.RouterGroup) PublicHandler {
	h := &handler{controller: controller, Group: group}

	group.GET(":lang", h.GetDefault)
	group.GET(":lang/:group", h.GetByGroup)

	swagger.SwaggerInfogeneral.BasePath = "/api/v1/i18n"

	return h
}

func (h *handler) GetDefault(ctx *gin.Context) {
	lang := ctx.Param("lang")
	i18n, err := h.controller.LoadDefaults(ctx, lang)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, i18n)
}

func (h *handler) GetByGroup(ctx *gin.Context) {
	lang := ctx.Param("lang")
	code := ctx.Param("group")
	i18n, err := h.controller.LoadByGroup(ctx, lang, code)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, i18n)
}
