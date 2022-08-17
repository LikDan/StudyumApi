package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/src/controllers"
	"studyum/src/dto"
	"studyum/src/utils"
	"time"
)

type UserHandler struct {
	IHandler

	controller controllers.IUserController

	Group *gin.RouterGroup
}

func NewUserHandler(authHandler IHandler, controller controllers.IUserController, group *gin.RouterGroup) *UserHandler {
	h := &UserHandler{IHandler: authHandler, controller: controller, Group: group}

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

	return h
}

func (u *UserHandler) putToken(ctx *gin.Context, token string) error {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:    "authToken",
		Value:   token,
		Path:    "/",
		Expires: time.Now().AddDate(1, 0, 0),
	})

	return nil
}

func (u *UserHandler) GetUser(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	ctx.JSON(http.StatusOK, user)
}

func (u *UserHandler) UpdateUser(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var data dto.UserSignUpData
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, err := u.controller.UpdateUser(ctx, user, data)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *UserHandler) LoginUser(ctx *gin.Context) {
	var data dto.UserLoginData
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, err := u.controller.LoginUser(ctx, data)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	if err = u.putToken(ctx, user.Token); err != nil {
		u.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *UserHandler) SignUpUser(ctx *gin.Context) {
	var data dto.UserSignUpData
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, err := u.controller.SignUpUser(ctx, data)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	if err = u.putToken(ctx, user.Token); err != nil {
		u.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *UserHandler) SignUpUserStage1(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var data dto.UserSignUpStage1Data
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, err := u.controller.SignUpUserStage1(ctx, user, data)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *UserHandler) OAuth2(ctx *gin.Context) {
	configName := ctx.Param("oauthConfigName")
	config := u.controller.GetOAuth2ConfigByName(configName)

	if config == nil {
		ctx.JSON(http.StatusBadRequest, "no such server")
		return
	}

	url := config.AuthCodeURL(ctx.Query("host"))
	ctx.Redirect(307, url)
}

func (u *UserHandler) CallbackOAuth2(ctx *gin.Context) {
	code := ctx.Query("code")

	user, err := u.controller.CallbackOAuth2(ctx, code)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.Redirect(307, "http://"+ctx.Query("state")+"/user/receiveToken?token="+user.Token)
}

func (u *UserHandler) PutAuthToken(ctx *gin.Context) {
	bytes, _ := ctx.GetRawData()
	token := string(bytes)

	if err := u.putToken(ctx, token); err != nil {
		u.Error(ctx, err)
	}

	user, err := u.controller.GetUserViaToken(ctx, token)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.JSON(200, user)
}

func (u *UserHandler) SignOutUser(ctx *gin.Context) {
	ctx.SetCookie("authToken", "", -1, "", "", false, false)
	ctx.JSON(200, "authToken")
}

func (u *UserHandler) RevokeToken(ctx *gin.Context) {
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
