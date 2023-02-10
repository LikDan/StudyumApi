package repositories

import (
	"context"
	"studyum/pkg/jwt/entities"
)

type Repository interface {
	Add(ctx context.Context, session entities.Session) error
	RemoveByID(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (entities.Session, error)
	Update(ctx context.Context, session entities.Session) error
	RemoveExpired(ctx context.Context) (int, error)
}
