package controllers

import (
	"context"
	"golang.org/x/oauth2"
	"studyum/src/models"
)

type IUserController interface {
	UpdateUser(ctx context.Context, user models.User, data models.UserSignUpData) (models.User, error)

	LoginUser(ctx context.Context, data models.UserLoginData) (models.User, error)
	SignUpUser(ctx context.Context, data models.UserSignUpData) (models.User, error)
	SignUpUserStage1(ctx context.Context, user models.User, data models.UserSignUpStage1Data) (models.User, error)

	UpdateToken(ctx context.Context, data models.UserLoginData, token string) error
	RevokeToken(ctx context.Context, token string) error

	GetUserViaToken(ctx context.Context, token string) (models.User, error)
	CallbackOAuth2(ctx context.Context, code string) (models.User, error)
	GetOAuth2ConfigByName(name string) *oauth2.Config
}
