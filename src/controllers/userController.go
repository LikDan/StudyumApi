package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"studyum/src/models"
	"studyum/src/repositories"
	"studyum/src/utils"
	"time"
)

type UserController struct {
	repository repositories.IUserRepository
}

func NewUserController(repository repositories.IUserRepository) *UserController {
	return &UserController{repository: repository}
}

func (u *UserController) putToken(ctx *gin.Context, user *models.User) *models.Error {
	if user.Token == "" {
		user.Token = utils.GenerateSecureToken()
		data := models.UserLoginData{
			Email:    user.Email,
			Password: user.Password,
		}

		if err := u.repository.UpdateToken(ctx, data, user.Token); err != nil {
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

func (u *UserController) SignUpUser(ctx context.Context, data models.UserSignUpData) (models.User, *models.Error) {
	if err := validator.New().Struct(&data); err != nil {
		return models.User{}, models.BindErrorStr("provide valid data", 400, models.UNDEFINED)
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
	if err := u.repository.SignUp(ctx, &user); err.Check() {
		return models.User{}, err
	}

	return user, models.EmptyError()
}

func (u *UserController) SignUpUserStage1(ctx context.Context, user models.User, data models.UserSignUpStage1Data) (models.User, *models.Error) {
	switch data.Type {
	case "group", "teacher":
		user.Type = data.Type
		user.StudyPlaceId = data.StudyPlaceId
		user.TypeName = data.TypeName
		break
	default:
		return models.User{}, models.BindErrorStr("provide valid data", 400, models.UNDEFINED)
	}

	if err := u.repository.SignUpStage1(ctx, &user); err.Check() {
		return models.User{}, err
	}

	return user, models.EmptyError()
}

func (u *UserController) UpdateUser(ctx context.Context, user models.User, data models.UserSignUpData) (models.User, *models.Error) {
	if err := validator.New().Struct(&data); err != nil {
		return models.User{}, models.BindErrorStr("provide valid data", 400, models.UNDEFINED)
	}

	if data.Password != "" && len(data.Password) > 8 {
		user.Password = utils.Hash(data.Password)
	}

	user.Login = data.Login
	user.Name = data.Name
	user.Email = data.Email
	if err := u.repository.UpdateUser(ctx, &user); err.Check() {
		return models.User{}, err
	}

	return user, models.EmptyError()
}

func (u *UserController) LoginUser(ctx context.Context, data models.UserLoginData) (models.User, *models.Error) {
	data.Password = utils.Hash(data.Password)

	var user models.User
	if err := u.repository.Login(ctx, &data, &user); err.Check() {
		return models.User{}, err
	}

	return user, models.EmptyError()
}

func (u *UserController) RevokeToken(ctx context.Context, token string) *models.Error {
	if err := u.repository.RevokeToken(ctx, token); err.Check() {
		return err
	}

	return models.EmptyError()
}
