package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"studyum/internal/auth/controllers"
	"studyum/internal/auth/dto"
)

type Auth struct {
	Middleware

	controller controllers.Auth

	Group *gin.RouterGroup
}

func NewAuth(middleware Middleware, controller controllers.Auth, group *gin.RouterGroup) *Auth {
	h := &Auth{Middleware: middleware, controller: controller, Group: group}

	group.PUT("login", h.Login)

	group.POST("signup", h.SignUp)
	group.PUT("signup/stage1", h.Auth(), h.SignUpUserStage1)
	group.POST("signup/code", h.Auth(), h.SignUpStage1ViaCode)
	group.DELETE("signout", h.Auth(), h.SignOut)

	group.POST("email/confirm", h.Auth(), h.ConfirmEmail)
	group.POST("email/resendCode", h.Auth(), h.ResendEmailCode)

	group.DELETE("sessions", h.Auth(), h.TerminateAllSessions)

	return h
}

func (h *Auth) Login(ctx *gin.Context) {
	var data dto.Login
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, pair, err := h.controller.Login(ctx, ctx.ClientIP(), data)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.SetTokenPairCookie(ctx, pair)

	ctx.JSON(http.StatusOK, user)
}

func (h *Auth) SignUp(ctx *gin.Context) {
	var data dto.SignUp
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, pair, err := h.controller.SignUp(ctx, ctx.ClientIP(), data)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.SetTokenPairCookie(ctx, pair)
	ctx.JSON(http.StatusOK, user)
}

func (h *Auth) SignUpStage1ViaCode(ctx *gin.Context) {
	user := h.GetUser(ctx)

	var data dto.SignUpWithCode
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.controller.SignUpStage1ViaCode(ctx, user, data.Code)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (h *Auth) SignUpUserStage1(ctx *gin.Context) {
	user := h.GetUser(ctx)

	var data dto.SignUpStage1
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.controller.SignUpStage1(ctx, user, data)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (h *Auth) SignOut(ctx *gin.Context) {
	token, _ := ctx.Cookie("refresh")
	if token != "" {
		if err := h.controller.SignOut(ctx, token); err != nil {
			if err != mongo.ErrNoDocuments {
				_ = ctx.Error(err)
				return
			}
		}
	}

	h.DeleteTokenPairCookie(ctx)
	ctx.Status(http.StatusNoContent)
}

func (h *Auth) ConfirmEmail(ctx *gin.Context) {
	user := h.GetUser(ctx)

	var data dto.VerificationCode
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := h.controller.ConfirmEmail(ctx, user, data); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *Auth) ResendEmailCode(ctx *gin.Context) {
	user := h.GetUser(ctx)
	if err := h.controller.ResendEmailCode(ctx, user); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *Auth) TerminateAllSessions(ctx *gin.Context) {
	user := h.GetUser(ctx)
	if err := h.controller.TerminateAll(ctx, user); err != nil {
		_ = ctx.Error(err)
		return
	}

	h.DeleteTokenPairCookie(ctx)
	ctx.Status(http.StatusNoContent)
}
