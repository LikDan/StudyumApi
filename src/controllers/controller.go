package controllers

import (
	"context"
	"studyum/src/models"
	"studyum/src/repositories"
)

type Controller struct {
	repository repositories.IUserRepository
}

func NewController(repository repositories.IUserRepository) *Controller {
	return &Controller{repository: repository}
}

func (a *Controller) Auth(ctx context.Context, token string) (models.User, error) {
	var user models.User
	err := a.repository.GetUserViaToken(ctx, token, &user)

	return user, err
}
