package controllers

import (
	"context"
	"studyum/src/entities"
)

type IController interface {
	Auth(ctx context.Context, token string) (entities.User, error)
}
