package controllers

import (
	"context"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/dto"
	"studyum/internal/entities"
	"studyum/internal/repositories"
	"studyum/internal/utils"
)

type UserController struct {
	repository repositories.UserRepository
}

func NewUserController(repository repositories.UserRepository) *UserController {
	return &UserController{repository: repository}
}

func (u *UserController) SignUpUser(ctx context.Context, data dto.UserSignUpData) (entities.User, error) {
	if err := validator.New().Struct(&data); err != nil {
		return entities.User{}, NotValidParams
	}

	data.Password = utils.Hash(data.Password)

	user := entities.User{
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
	if _, err := u.repository.SignUp(ctx, user); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (u *UserController) SignUpUserStage1(ctx context.Context, user entities.User, data dto.UserSignUpStage1Data) (entities.User, error) {
	switch data.Type {
	case "group", "teacher":
		user.Type = data.Type
		user.StudyPlaceId = data.StudyPlaceId
		user.TypeName = data.TypeName
		break
	default:
		return entities.User{}, NotValidParams
	}

	if err := u.repository.SignUpStage1(ctx, user); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (u *UserController) UpdateUser(ctx context.Context, user entities.User, data dto.UserSignUpData) (entities.User, error) {
	if err := validator.New().Struct(&data); err != nil {
		return entities.User{}, NotValidParams
	}

	if data.Password != "" && len(data.Password) > 8 {
		user.Password = utils.Hash(data.Password)
	}

	user.Login = data.Login
	user.Name = data.Name
	user.Email = data.Email
	if err := u.repository.UpdateUser(ctx, user); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (u *UserController) LoginUser(ctx context.Context, data dto.UserLoginData) (entities.User, error) {
	data.Password = utils.Hash(data.Password)

	var user entities.User
	if _, err := u.repository.Login(ctx, "", ""); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (u *UserController) UpdateTokenByID(ctx context.Context, id primitive.ObjectID, token string) error {
	return u.repository.UpdateToken(ctx, id, token)
}

func (u *UserController) RevokeToken(ctx context.Context, token string) error {
	return u.repository.RevokeToken(ctx, token)
}
