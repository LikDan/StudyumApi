package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"path/filepath"
	"studyum/internal/controllers"
	"studyum/internal/dto"
	"studyum/internal/utils"
)

type UserHandler interface {
	GetUser(ctx *gin.Context)
	UpdateUser(ctx *gin.Context)

	PutPicture(ctx *gin.Context)

	LoginUser(ctx *gin.Context)
	SignUpUser(ctx *gin.Context)
	SignUpUserStage1(ctx *gin.Context)
	SignUpUserWithToken(ctx *gin.Context)
	SignOutUser(ctx *gin.Context)

	OAuth2(ctx *gin.Context)
	CallbackOAuth2(ctx *gin.Context)

	PutFirebaseToken(ctx *gin.Context)

	RevokeToken(ctx *gin.Context)
	TerminateSession(ctx *gin.Context)

	GetAccept(ctx *gin.Context)
	Accept(ctx *gin.Context)
	Block(ctx *gin.Context)
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
	group.POST("signup/withToken", h.SignUpUserWithToken)
	group.POST("signup", h.SignUpUser)
	group.PUT("signup/stage1", h.AuthBlockedOrNotAccepted(), h.SignUpUserStage1)

	group.GET("auth/:oauth", h.OAuth2)
	group.GET("oauth2/callback/:oauth", h.CallbackOAuth2)

	group.DELETE("signout", h.Auth(), h.SignOutUser)
	group.DELETE("revoke", h.Auth(), h.RevokeToken)
	group.DELETE("terminate/:ip", h.Auth(), h.TerminateSession)

	group.POST("code", h.Auth("manageUsers"), h.CreateCode)
	group.GET("accept", h.Auth("manageUsers"), h.GetAccept)
	group.POST("accept", h.Auth("manageUsers"), h.Accept)
	group.POST("block", h.Auth("manageUsers"), h.Block)

	group.PUT("firebase/token", h.Auth(), h.PutFirebaseToken)

	group.POST("files/profile", h.PutPicture)

	return h
}

func (u *userHandler) GetUser(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	ctx.JSON(http.StatusOK, user)
}

func (u *userHandler) PutPicture(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		return
	}

	filename := filepath.Base(file.Filename)
	if err := c.SaveUploadedFile(file, filename); err != nil {
		return
	}
}

func (u *userHandler) UpdateUser(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var data dto.EditUserDTO
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, pair, err := u.controller.UpdateUser(ctx, user, data)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	u.SetTokenPairCookie(ctx, pair)
	ctx.JSON(http.StatusOK, user)
}

func (u *userHandler) LoginUser(ctx *gin.Context) {
	var data dto.UserLoginDTO
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, pair, err := u.controller.LoginUser(ctx, data, ctx.ClientIP())
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

	user, pair, err := u.controller.SignUpUser(ctx, data, ctx.ClientIP())
	if err != nil {
		u.Error(ctx, err)
		return
	}

	u.SetTokenPairCookie(ctx, pair)
	ctx.JSON(http.StatusOK, user)
}

func (u *userHandler) SignUpUserWithToken(ctx *gin.Context) {
	var data dto.UserSignUpWithCodeDTO
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, pair, err := u.controller.SignUpUserWithCode(ctx, ctx.ClientIP(), data)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	u.SetTokenPairCookie(ctx, pair)

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
	configName := ctx.Param("oauth")
	code := ctx.Query("code")

	pair, err := u.controller.CallbackOAuth2(ctx, configName, code)
	if err != nil {
		u.Error(ctx, err)
		return
	}

	u.SetTokenPairCookie(ctx, pair)

	ctx.SetSameSite(http.SameSiteNoneMode)
	ctx.Redirect(302, "http://"+ctx.Query("state")+"/")
}

func (u *userHandler) PutFirebaseToken(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var token string
	if err := ctx.BindJSON(&token); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := u.controller.PutFirebaseTokenByUserID(ctx, user.Id, token); err != nil {
		u.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *userHandler) SignOutUser(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "no cookie")
	}

	if err = u.controller.SignOut(ctx, refreshToken); err != nil {
		u.Error(ctx, err)
	}

	ctx.SetCookie("access", "", -1, "", "", false, false)
	ctx.SetCookie("refresh", "", -1, "", "", false, false)

	ctx.JSON(http.StatusOK, "successful")
}

func (u *userHandler) RevokeToken(ctx *gin.Context) {
	token, err := ctx.Cookie("refresh")
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

func (u *userHandler) TerminateSession(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)
	ip := ctx.Param("ip")

	err := u.controller.TerminateSession(ctx, user, ip)
	if err != nil {
		u.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, ip)
}

func (u *userHandler) CreateCode(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var data dto.UserCreateCodeDTO
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	code, err := u.controller.CreateCode(ctx, user, data)
	if err != nil {
		u.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, code)
}

func (u *userHandler) GetAccept(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	users, err := u.controller.GetAccept(ctx, user)
	if err != nil {
		u.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, users)
}

func (u *userHandler) Accept(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

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

	if err = u.controller.Accept(ctx, user, id); err != nil {
		u.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, id)
}

func (u *userHandler) Block(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

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

	if err = u.controller.Block(ctx, user, id); err != nil {
		u.Error(ctx, err)
	}

	ctx.JSON(http.StatusOK, id)
}
