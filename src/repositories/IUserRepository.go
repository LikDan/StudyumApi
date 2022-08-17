package repositories

import (
	"context"
	"studyum/src/dto"
	"studyum/src/entities"
)

type IUserRepository interface {
	GetUserViaToken(ctx context.Context, token string, user *entities.User) error
	GetUserByEmail(ctx context.Context, email string, user *entities.User) error

	SignUp(ctx context.Context, user *entities.User) error
	SignUpStage1(ctx context.Context, user *entities.User) error

	Login(ctx context.Context, data *dto.UserLoginData, user *entities.User) error

	UpdateUser(ctx context.Context, user *entities.User) error

	RevokeToken(ctx context.Context, token string) error
	UpdateToken(ctx context.Context, data dto.UserLoginData, token string) error
	UpdateUserTokenByEmail(ctx context.Context, email, token string) error
}
