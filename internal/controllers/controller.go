package controllers

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/entities"
	"studyum/internal/repositories"
)

var NotAuthorizationError = errors.New("not authorized")

type Controller struct {
	repository repositories.UserRepository
}

func NewController(repository repositories.UserRepository) *Controller {
	return &Controller{repository: repository}
}

func (a *Controller) Auth(ctx context.Context, token string) (entities.User, error) {
	var user entities.User
	_, err := a.repository.GetUserViaToken(ctx, token)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return entities.User{}, NotAuthorizationError
		} else {
			return entities.User{}, err
		}
	}

	return user, err
}
