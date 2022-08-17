package repositories

import (
	"context"
	"studyum/src/models"
)

type IUserRepository interface {
	GetUserViaToken(ctx context.Context, token string, user *models.User) error
	GetUserByEmail(ctx context.Context, email string, user *models.User) error

	SignUp(ctx context.Context, user *models.User) error
	SignUpStage1(ctx context.Context, user *models.User) error

	Login(ctx context.Context, data *models.UserLoginData, user *models.User) error

	UpdateUser(ctx context.Context, user *models.User) error

	RevokeToken(ctx context.Context, token string) error
	UpdateToken(ctx context.Context, data models.UserLoginData, token string) error
	UpdateUserTokenByEmail(ctx context.Context, email, token string) error
}
