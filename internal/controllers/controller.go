package controllers

import (
	"context"
	"studyum/internal/entities"
	"studyum/internal/repositories"
)

type Controller struct {
	repository repositories.IUserRepository
}

func NewController(repository repositories.IUserRepository) *Controller {
	return &Controller{repository: repository}
}

func (a *Controller) Auth(ctx context.Context, token string) (entities.User, error) {
	var user entities.User
	err := a.repository.GetUserViaToken(ctx, token, &user)

	return user, err
}
