package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/internal/auth/controllers"
	"studyum/internal/global"
)

type OAuth2 struct {
	global.Handler
	Middleware

	controller controllers.OAuth2

	Group *gin.RouterGroup
}

func NewOAuth2(handler global.Handler, middleware Middleware, controller controllers.OAuth2, group *gin.RouterGroup) *OAuth2 {
	h := &OAuth2{Handler: handler, Middleware: middleware, controller: controller, Group: group}

	group.GET(":service", h.Auth)
	group.GET("/callback/:service", h.Receive)
	group.POST("/token", h.SetToken)

	return h
}

func (h *OAuth2) Auth(ctx *gin.Context) {
	service := ctx.Param("service")
	redirectHost := ctx.Query("redirect")

	url, err := h.controller.GetServiceURL(ctx, service, redirectHost)
	if err != nil {
		h.Error(ctx, err)
		return
	}

	ctx.Redirect(307, url)
}

func (h *OAuth2) Receive(ctx *gin.Context) {
	service := ctx.Param("service")
	code := ctx.Query("code")

	pair, err := h.controller.ReceiveUser(ctx, service, code)
	if err != nil {
		h.Error(ctx, err)
		return
	}

	ctx.Redirect(http.StatusPermanentRedirect, ctx.Query("state")+"/?token="+pair.Refresh)
}

func (h *OAuth2) SetToken(ctx *gin.Context) {
	var data struct {
		Token string `json:"token"`
	}

	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.Set("refresh", data.Token)
	h.Middleware.Auth()(ctx)

	user := h.GetUser(ctx)
	user = h.controller.DecryptUser(ctx, user)

	ctx.JSON(http.StatusOK, user)
}