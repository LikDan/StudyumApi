package controllers

import (
	"context"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
	"studyum/src/repositories"
	"studyum/src/utils"
)

type UserController struct {
	repository repositories.IUserRepository
}

func NewUserController(repository repositories.IUserRepository) *UserController {
	return &UserController{repository: repository}
}

func (u *UserController) SignUpUser(ctx context.Context, data models.UserSignUpData) (models.User, error) {
	if err := validator.New().Struct(&data); err != nil {
		return models.User{}, NotValidParams
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
	if err := u.repository.SignUp(ctx, &user); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (u *UserController) SignUpUserStage1(ctx context.Context, user models.User, data models.UserSignUpStage1Data) (models.User, error) {
	switch data.Type {
	case "group", "teacher":
		user.Type = data.Type
		user.StudyPlaceId = data.StudyPlaceId
		user.TypeName = data.TypeName
		break
	default:
		return models.User{}, NotValidParams
	}

	if err := u.repository.SignUpStage1(ctx, &user); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (u *UserController) UpdateUser(ctx context.Context, user models.User, data models.UserSignUpData) (models.User, error) {
	if err := validator.New().Struct(&data); err != nil {
		return models.User{}, NotValidParams
	}

	if data.Password != "" && len(data.Password) > 8 {
		user.Password = utils.Hash(data.Password)
	}

	user.Login = data.Login
	user.Name = data.Name
	user.Email = data.Email
	if err := u.repository.UpdateUser(ctx, &user); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (u *UserController) LoginUser(ctx context.Context, data models.UserLoginData) (models.User, error) {
	data.Password = utils.Hash(data.Password)

	var user models.User
	if err := u.repository.Login(ctx, &data, &user); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (u *UserController) UpdateToken(ctx context.Context, data models.UserLoginData, token string) error {
	return u.repository.UpdateToken(ctx, data, token)
}

func (u *UserController) RevokeToken(ctx context.Context, token string) error {
	return u.repository.RevokeToken(ctx, token)
}
