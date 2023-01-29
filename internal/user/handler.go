package user

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	auth "studyum/internal/auth/handlers"
	"studyum/internal/global"
)

type Handler interface {
	GetUser(ctx *gin.Context)
	UpdateUser(ctx *gin.Context)

	PutFirebaseToken(ctx *gin.Context)

	GetAccept(ctx *gin.Context)
	Accept(ctx *gin.Context)
	Block(ctx *gin.Context)
}

type handler struct {
	global.Handler
	auth.Middleware

	controller Controller

	Group *gin.RouterGroup
}

func NewUserHandler(authHandler global.Handler, middleware auth.Middleware, controller Controller, group *gin.RouterGroup) Handler {
	h := &handler{Handler: authHandler, Middleware: middleware, controller: controller, Group: group}

	group.GET("", h.Auth(), h.GetUser)
	group.PUT("", h.Auth(), h.UpdateUser)

	group.GET("accept", h.MemberAuth("manageUsers"), h.GetAccept)
	group.POST("accept", h.MemberAuth("manageUsers"), h.Accept)
	group.POST("block", h.MemberAuth("manageUsers"), h.Block)

	group.PUT("firebase/token", h.Auth(), h.PutFirebaseToken)

	return h
}

func (h *handler) GetUser(ctx *gin.Context) {
	user := h.Handler.GetUser(ctx)
	user = h.controller.DecryptUser(ctx, user)

	ctx.JSON(http.StatusOK, user)
}

func (h *handler) UpdateUser(ctx *gin.Context) {
	user := h.Handler.GetUser(ctx)

	var data Edit
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	token, _ := ctx.Cookie("refresh")
	user, pair, err := h.controller.UpdateUser(ctx, user, token, ctx.ClientIP(), data)
	if err != nil {
		h.Error(ctx, err)
		return
	}

	h.SetTokenPairCookie(ctx, pair)
	ctx.JSON(http.StatusOK, user)
}

func (h *handler) PutFirebaseToken(ctx *gin.Context) {
	user := h.Handler.GetUser(ctx)

	var token string
	if err := ctx.BindJSON(&token); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := h.controller.PutFirebaseTokenByUserID(ctx, user.Id, token); err != nil {
		h.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (h *handler) GetAccept(ctx *gin.Context) {
	user := h.Handler.GetUser(ctx)

	users, err := h.controller.GetAccept(ctx, user)
	if err != nil {
		h.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, users)
}

func (h *handler) Accept(ctx *gin.Context) {
	user := h.Handler.GetUser(ctx)

	var idHex string
	if err := ctx.BindJSON(&idHex); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = h.controller.Accept(ctx, user, id); err != nil {
		h.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, id)
}

func (h *handler) Block(ctx *gin.Context) {
	user := h.Handler.GetUser(ctx)

	var idHex string
	if err := ctx.BindJSON(&idHex); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = h.controller.Block(ctx, user, id); err != nil {
		h.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, id)
}
