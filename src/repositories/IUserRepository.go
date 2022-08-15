package repositories

import (
	"context"
	"studyum/src/models"
)

type IUserRepository interface {
	GetUserViaToken(ctx context.Context, token string, user *models.User) *models.Error
	GetUserByEmail(ctx context.Context, email string, user *models.User) *models.Error

	SignUp(ctx context.Context, user *models.User) *models.Error
	SignUpStage1(ctx context.Context, user *models.User) *models.Error

	Login(ctx context.Context, data *models.UserLoginData, user *models.User) *models.Error

	UpdateUser(ctx context.Context, user *models.User) *models.Error

	RevokeToken(ctx context.Context, token string) *models.Error
	UpdateToken(ctx context.Context, data models.UserLoginData, token string) *models.Error
	UpdateUserTokenByEmail(ctx context.Context, email, token string) *models.Error
}
