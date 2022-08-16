package controllers

import (
	"context"
	"studyum/src/models"
	"studyum/src/repositories"
)

type AuthController struct {
	repository repositories.IUserRepository
}

func NewAuthController(repository repositories.IUserRepository) *AuthController {
	return &AuthController{repository: repository}
}

func (a *AuthController) Auth(ctx context.Context, token string) (models.User, *models.Error) {
	var user models.User
	err := a.repository.GetUserViaToken(ctx, token, &user)

	return user, err
}
