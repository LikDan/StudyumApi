package controllers

import (
	"context"
	"golang.org/x/oauth2"
	"studyum/src/dto"
	"studyum/src/entities"
)

type IUserController interface {
	UpdateUser(ctx context.Context, user entities.User, data dto.UserSignUpData) (entities.User, error)

	LoginUser(ctx context.Context, data dto.UserLoginData) (entities.User, error)
	SignUpUser(ctx context.Context, data dto.UserSignUpData) (entities.User, error)
	SignUpUserStage1(ctx context.Context, user entities.User, data dto.UserSignUpStage1Data) (entities.User, error)

	UpdateToken(ctx context.Context, data dto.UserLoginData, token string) error
	RevokeToken(ctx context.Context, token string) error

	GetUserViaToken(ctx context.Context, token string) (entities.User, error)
	CallbackOAuth2(ctx context.Context, code string) (entities.User, error)
	GetOAuth2ConfigByName(name string) *oauth2.Config
}
