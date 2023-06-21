package repositories

import (
	"context"
	"github.com/pkg/errors"
	"studyum/pkg/jwt/entities"
)

var NotValidRefreshTokenErr = errors.New("Not valid refresh token")

type Repository interface {
	Add(ctx context.Context, session entities.Session) error
	RemoveByID(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (entities.Session, error)
	Update(ctx context.Context, session entities.Session) error
	RemoveExpired(ctx context.Context) (int, error)
}
