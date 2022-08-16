package controllers

import (
	"context"
	"golang.org/x/oauth2"
	"studyum/src/models"
)

type IUserController interface {
	UpdateUser(ctx context.Context, user models.User, data models.UserSignUpData) (models.User, *models.Error)

	LoginUser(ctx context.Context, data models.UserLoginData) (models.User, *models.Error)
	SignUpUser(ctx context.Context, data models.UserSignUpData) (models.User, *models.Error)
	SignUpUserStage1(ctx context.Context, user models.User, data models.UserSignUpStage1Data) (models.User, *models.Error)

	RevokeToken(ctx context.Context, token string) *models.Error

	GetUserViaToken(ctx context.Context, token string) (models.User, *models.Error)
	CallbackOAuth2(ctx context.Context, code string) (models.User, *models.Error)
	GetOAuth2ConfigByName(name string) *oauth2.Config
}
