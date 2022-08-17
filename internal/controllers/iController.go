package controllers

import (
	"context"
	"studyum/internal/entities"
)

type IController interface {
	Auth(ctx context.Context, token string) (entities.User, error)
}
