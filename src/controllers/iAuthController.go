package controllers

import (
	"context"
	"studyum/src/models"
)

type IAuthController interface {
	Auth(ctx context.Context, token string) (models.User, *models.Error)
}
