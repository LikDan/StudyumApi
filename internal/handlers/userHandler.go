package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/internal/controllers"
	"studyum/internal/dto"
	"studyum/internal/utils"
)

type UserHandler interface {
	GetUser(ctx *gin.Context)
	UpdateUser(ctx *gin.Context)

	LoginUser(ctx *gin.Context)
	SignUpUser(ctx *gin.Context)
	SignUpUserStage1(ctx *gin.Context)
	SignOutUser(ctx *gin.Context)

	OAuth2(ctx *gin.Context)
	CallbackOAuth2(ctx *gin.Context)
	PutAuthToken(ctx *gin.Context)

	PutFirebaseToken(ctx *gin.Context)

	RevokeToken(ctx *gin.Context)
}

type userHandler struct {
	Handler

	controller controllers.UserController

	Group *gin.RouterGroup
}

func NewUserHandler(authHandler Handler, controller controllers.UserController, group *gin.RouterGroup) UserHandler {
	h := &userHandler{Handler: authHandler, controller: controller, Group: group}

	group.GET("", h.Auth(), h.GetUser)
	group.PUT("", h.Auth(), h.UpdateUser)

	group.PUT("login", h.LoginUser)
	group.POST("signup", h.SignUpUser)

	group.PUT("signup/stage1", h.Auth(), h.SignUpUserStage1)

	group.GET("auth/:oauth", h.OAuth2)
	group.GET("callback", h.CallbackOAuth2)
	group.PUT("auth/token", h.PutAuthToken)

	group.DELETE("signout", h.SignOutUser)
	group.DELETE("revoke", h.RevokeToken)

	group.PUT("firebase/token", h.Auth(), h.PutFirebaseToken)

	return h
}

func (u *userHandler) GetUser(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	ctx.JSON(http.StatusOK, user)
}

func (u *userHandler) UpdateUser(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var data dto.EditUserDTO
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	pair, err := u.controller.UpdateUser(ctx, user, data)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, pair)
}

func (u *userHandler) LoginUser(ctx *gin.Context) {
	var data dto.UserLoginDTO
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, pair, err := u.controller.LoginUser(ctx, data)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	u.SetTokenPairCookie(ctx, pair)

	ctx.JSON(http.StatusOK, user)
}

func (u *userHandler) SignUpUser(ctx *gin.Context) {
	var data dto.UserSignUpDTO
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := u.controller.SignUpUser(ctx, data)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *userHandler) SignUpUserStage1(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var data dto.UserSignUpStage1DTO
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := u.controller.SignUpUserStage1(ctx, user, data)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *userHandler) OAuth2(ctx *gin.Context) {
	configName := ctx.Param("oauth")
	config := u.controller.GetOAuth2ConfigByName(configName)

	if config == nil {
		ctx.JSON(http.StatusBadRequest, "no such server")
		return
	}

	url := config.AuthCodeURL(ctx.Query("host"))
	ctx.Redirect(307, url)
}

func (u *userHandler) CallbackOAuth2(ctx *gin.Context) {
	code := ctx.Query("code")

	user, err := u.controller.CallbackOAuth2(ctx, code)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.Redirect(307, "http://"+ctx.Query("state")+"/user/receiveToken?token="+user.Token)
}

func (u *userHandler) PutAuthToken(ctx *gin.Context) {
	bytes, _ := ctx.GetRawData()
	token := string(bytes)

	user, err := u.controller.GetUserViaToken(ctx, token)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *userHandler) PutFirebaseToken(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var token string
	if err := ctx.BindJSON(&token); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := u.controller.PutFirebaseToken(ctx, user.Token, token); err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *userHandler) SignOutUser(ctx *gin.Context) {
	ctx.SetCookie("access", "", -1, "", "", false, false)
	ctx.SetCookie("refresh", "", -1, "", "", false, false)

	ctx.JSON(http.StatusOK, "successful")
}

func (u *userHandler) RevokeToken(ctx *gin.Context) {
	token, err := ctx.Cookie("authToken")
	if err != nil {
		ctx.JSON(http.StatusForbidden, "token not present")
		return
	}

	err = u.controller.RevokeToken(ctx, token)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, token)
}
