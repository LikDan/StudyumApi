package controllers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	utils "studyum/src/api"
	"studyum/src/db"
	"studyum/src/models"
	"time"
)

func AuthUserViaToken(token string, user *models.User, permissions ...string) *models.Error {
	var user_ models.User
	if err := db.GetUserViaToken(token, &user_); err.Error != nil {
		return err
	}

	for _, permission := range permissions {
		if !utils.SliceContains(user.Permissions, permission) {
			return models.BindErrorStr("no permission(s)", 403, utils.UNDEFINED)
		}
	}

	*user = user_
	return models.EmptyError()
}

func AuthUserViaContext(ctx *gin.Context, user *models.User, permissions ...string) *models.Error {
	token, err := ctx.Cookie("authToken")
	if err != nil {
		return models.BindError(err, 401, utils.UNDEFINED)
	}

	AuthUserViaToken(token, user, permissions...)

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

		if _, err := db.UsersCollection.UpdateOne(ctx, data, bson.M{"$set": bson.M{"token": user.Token}}); err != nil {
			return models.BindError(err, 418, utils.WARNING)
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
	if err := ctx.BindJSON(&data); models.BindError(err, 400, utils.UNDEFINED).CheckAndResponse(ctx) {
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
	if err := db.SignUp(&user); err.CheckAndResponse(ctx) {
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
	if err := ctx.BindJSON(&data); models.BindError(err, 400, utils.UNDEFINED).CheckAndResponse(ctx) {
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

	if err := db.SignUpStage1(&user); err.CheckAndResponse(ctx) {
		return
	}

	if err := putToken(ctx, &user); err.CheckAndResponse(ctx) {
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
	if err := ctx.BindJSON(&data); models.BindError(err, 400, utils.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	data.Password = utils.Hash(data.Password)

	var user models.User
	if err := db.Login(&data, &user); err.CheckAndResponse(ctx) {
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
		models.BindErrorStr("not authorized", 401, utils.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	if db.RevokeToken(token).CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, token)
}
