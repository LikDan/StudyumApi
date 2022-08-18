package controllers

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"studyum/internal/dto"
	"studyum/internal/entities"
)

type IUserController interface {
	UpdateUser(ctx context.Context, user entities.User, data dto.UserSignUpData) (entities.User, error)

	LoginUser(ctx context.Context, data dto.UserLoginData) (entities.User, error)
	SignUpUser(ctx context.Context, data dto.UserSignUpData) (entities.User, error)
	SignUpUserStage1(ctx context.Context, user entities.User, data dto.UserSignUpStage1Data) (entities.User, error)

	UpdateTokenByID(ctx context.Context, id primitive.ObjectID, token string) error
	RevokeToken(ctx context.Context, token string) error

	GetUserViaToken(ctx context.Context, token string) (entities.User, error)
	CallbackOAuth2(ctx context.Context, code string) (entities.User, error)
	GetOAuth2ConfigByName(name string) *oauth2.Config
}
