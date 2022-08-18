package controllers

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/entities"
	"studyum/internal/repositories"
)

var NotAuthorizationError = errors.New("not authorized")

type Controller interface {
	Auth(ctx context.Context, token string, permissions ...string) (entities.User, error)
}

type controller struct {
	repository repositories.UserRepository
}

func NewController(repository repositories.UserRepository) Controller {
	return &controller{repository: repository}
}

func (a *controller) Auth(ctx context.Context, token string, permissions ...string) (entities.User, error) {
	var user entities.User
	_, err := a.repository.GetUserViaToken(ctx, token, permissions...)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return entities.User{}, NotAuthorizationError
		} else {
			return entities.User{}, err
		}
	}

	return user, err
}
