package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	_ "github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"studyum/src/models"
	"studyum/src/repositories"
	"studyum/src/utils"
	"time"
)

var UserRepository repositories.IUserRepository

func AuthUserViaToken(ctx context.Context, token string, user *models.User, permissions ...string) *models.Error {
	var user_ models.User
	if err := UserRepository.GetUserViaToken(ctx, token, &user_); err.Error != nil {
		return err
	}

	for _, permission := range permissions {
		if !utils.SliceContains(user_.Permissions, permission) {
			return models.BindErrorStr("no permission(s)", 403, models.UNDEFINED)
		}
	}

	*user = user_
	return models.EmptyError()
}

func AuthUserViaContext(ctx *gin.Context, user *models.User, permissions ...string) *models.Error {
	token, err := ctx.Cookie("authToken")
	if err != nil {
		return models.BindError(err, 401, models.UNDEFINED)
	}

	if err := AuthUserViaToken(ctx, token, user, permissions...); err.Check() {
		return err
	}

	return models.EmptyError()
}

func GetUser(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, user)
}

func putToken(ctx *gin.Context, user *models.User) *models.Error {
	if user.Token == "" {
		user.Token = utils.GenerateSecureToken()
		data := models.UserLoginData{
			Email:    user.Email,
			Password: user.Password,
		}

		if err := UserRepository.UpdateToken(ctx, data, user.Token); err != nil {
			return err
		}
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:    "authToken",
		Value:   user.Token,
		Path:    "/",
		Expires: time.Now().AddDate(1, 0, 0),
	})

	return models.EmptyError()
}

func SignUpUser(ctx *gin.Context) {
	var data models.UserSignUpData
	if err := ctx.BindJSON(&data); models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	if err := validator.New().Struct(&data); err != nil {
		models.BindErrorStr("provide valid data", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	data.Password = utils.Hash(data.Password)

	user := models.User{
		Id:            primitive.NilObjectID,
		Token:         "",
		Password:      data.Password,
		Email:         data.Email,
		VerifiedEmail: false,
		Login:         data.Login,
		Name:          data.Name,
		PictureUrl:    "https://www.shareicon.net/data/128x128/2016/07/05/791214_man_512x512.png",
		Type:          "",
		TypeName:      "",
		StudyPlaceId:  0,
		Permissions:   nil,
		Accepted:      false,
		Blocked:       false,
	}
	if err := UserRepository.SignUp(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	if err := putToken(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, user)
}

func SignUpUserStage1(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	var data models.UserSignUpStage1Data
	if err := ctx.BindJSON(&data); models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	if data.Type == "group" {
		user.Type = data.Type
		user.StudyPlaceId = data.StudyPlaceId
		user.TypeName = data.TypeName
	} else if data.Type == "teacher" {
		user.Type = data.Type
		user.StudyPlaceId = data.StudyPlaceId
		user.TypeName = user.Name
	} else {
		ctx.JSON(400, "provide valid data")
		return
	}

	if err := UserRepository.SignUpStage1(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	if err := putToken(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, user)
}

func UpdateUser(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	var data models.UserSignUpData
	if err := ctx.BindJSON(&data); models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	if err := validator.New().Struct(&data); err != nil {
		models.BindErrorStr("provide valid data", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	if data.Password != "" && len(data.Password) > 8 {
		user.Password = utils.Hash(data.Password)
	}

	user.Login = data.Login
	user.Name = data.Name
	user.Email = data.Email
	if err := UserRepository.UpdateUser(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, user)
}

func SignOutUser(ctx *gin.Context) {
	ctx.SetCookie("authToken", "", -1, "", "", false, false)
	ctx.JSON(200, "authToken")
}

func LoginUser(ctx *gin.Context) {
	var data models.UserLoginData
	if err := ctx.BindJSON(&data); models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	data.Password = utils.Hash(data.Password)

	var user models.User
	if err := UserRepository.Login(ctx, &data, &user); err.CheckAndResponse(ctx) {
		return
	}

	if err := putToken(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, user)
}

func RevokeToken(ctx *gin.Context) {
	token, err := ctx.Cookie("authToken")
	if err != nil {
		models.BindErrorStr("not authorized", 401, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	if UserRepository.RevokeToken(ctx, token).CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, token)
}
