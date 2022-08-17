package controllers

import (
	"context"
	"studyum/src/models"
)

type IController interface {
	Auth(ctx context.Context, token string) (models.User, error)
}
