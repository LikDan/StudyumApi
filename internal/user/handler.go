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

	RevokeToken(ctx *gin.Context)
	TerminateSession(ctx *gin.Context)

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

	group.DELETE("signout", h.Auth(), h.SignOutUser)
	group.DELETE("revoke", h.Auth(), h.RevokeToken)
	group.DELETE("terminate/:ip", h.Auth(), h.TerminateSession)

	group.POST("code", h.MemberAuth("manageUsers"), h.CreateCode)
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

func (h *handler) CallbackOAuth2(ctx *gin.Context) {
	//configName := ctx.Param("oauth")
	//code := ctx.Query("code")
	//
	//pair, err := h.controller.CallbackOAuth2(ctx, configName, code)
	//if err != nil {
	//	h.Error(ctx, err)
	//	return
	//}
	//
	//h.SetTokenPairCookie(ctx, pair)

	ctx.SetSameSite(http.SameSiteNoneMode)
	ctx.Redirect(302, "http://"+ctx.Query("state")+"/")
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

func (h *handler) SignOutUser(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "no cookie")
	}

	if err = h.controller.SignOut(ctx, refreshToken); err != nil {
		h.Error(ctx, err)
	}

	ctx.SetCookie("access", "", -1, "", "", false, false)
	ctx.SetCookie("refresh", "", -1, "", "", false, false)

	ctx.JSON(http.StatusOK, "successful")
}

func (h *handler) RevokeToken(ctx *gin.Context) {
	token, err := ctx.Cookie("refresh")
	if err != nil {
		ctx.JSON(http.StatusForbidden, "token not present")
		return
	}

	err = h.controller.RevokeToken(ctx, token)
	if err != nil {
		h.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, token)
}

func (h *handler) TerminateSession(ctx *gin.Context) {
	user := h.Handler.GetUser(ctx)
	ip := ctx.Param("ip")

	err := h.controller.TerminateSession(ctx, user, ip)
	if err != nil {
		h.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, ip)
}

func (h *handler) CreateCode(ctx *gin.Context) {
	user := h.Handler.GetUser(ctx)

	var data UserCreateCodeDTO
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	code, err := h.controller.CreateCode(ctx, user, data)
	if err != nil {
		h.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, code)
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
