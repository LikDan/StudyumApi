package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	auth "studyum/internal/auth/handlers"
	"studyum/internal/user/controllers"
	"studyum/internal/user/dto"
)

type Handler interface {
	GetUser(ctx *gin.Context)
	UpdateUser(ctx *gin.Context)

	PutFirebaseToken(ctx *gin.Context)

	GetAccept(ctx *gin.Context)
	Accept(ctx *gin.Context)
	Block(ctx *gin.Context)

	ResetPassword(ctx *gin.Context)
	ResetPasswordViaCode(ctx *gin.Context)
}

type handler struct {
	auth.Middleware

	controller controllers.Controller

	Group *gin.RouterGroup
}

func NewUserHandler(middleware auth.Middleware, controller controllers.Controller, group *gin.RouterGroup) Handler {
	h := &handler{Middleware: middleware, controller: controller, Group: group}

	group.GET("", h.Auth(), h.GetUser)
	group.PUT("", h.Auth(), h.UpdateUser)

	group.GET("accept", h.MemberAuth("manageUsers"), h.GetAccept)
	group.POST("accept", h.MemberAuth("manageUsers"), h.Accept)
	group.POST("block", h.MemberAuth("manageUsers"), h.Block)

	group.PUT("firebase/token", h.Auth(), h.PutFirebaseToken)

	group.POST("password/reset", h.ResetPassword)
	group.PUT("password/reset", h.ResetPasswordViaCode)

	return h
}

func (h *handler) GetUser(ctx *gin.Context) {
	user := h.Middleware.GetUser(ctx)
	user = h.controller.DecryptUser(ctx, user)

	ctx.JSON(http.StatusOK, user)
}

func (h *handler) UpdateUser(ctx *gin.Context) {
	user := h.Middleware.GetUser(ctx)

	var data dto.Edit
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	token, _ := ctx.Cookie("refresh")
	user, pair, err := h.controller.UpdateUser(ctx, user, token, ctx.ClientIP(), data)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.SetTokenPairCookie(ctx, pair)
	ctx.JSON(http.StatusOK, user)
}

func (h *handler) PutFirebaseToken(ctx *gin.Context) {
	user := h.Middleware.GetUser(ctx)

	var token string
	if err := ctx.BindJSON(&token); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := h.controller.PutFirebaseTokenByUserID(ctx, user.Id, token); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (h *handler) GetAccept(ctx *gin.Context) {
	user := h.Middleware.GetUser(ctx)

	users, err := h.controller.GetAccept(ctx, user)
	if err != nil {
		_ = ctx.Error(err)
	}

	ctx.JSON(http.StatusOK, users)
}

func (h *handler) Accept(ctx *gin.Context) {
	user := h.Middleware.GetUser(ctx)

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
		_ = ctx.Error(err)
	}

	ctx.JSON(http.StatusOK, id)
}

func (h *handler) Block(ctx *gin.Context) {
	user := h.Middleware.GetUser(ctx)

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
		_ = ctx.Error(err)
	}

	ctx.JSON(http.StatusOK, id)
}

func (h *handler) ResetPassword(ctx *gin.Context) {
	var data struct {
		Email string `json:"email"`
	}
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	err := h.controller.RecoverPassword(ctx, data.Email)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *handler) ResetPasswordViaCode(ctx *gin.Context) {
	var data dto.ResetPassword
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	err := h.controller.ResetPasswordViaCode(ctx, data)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
