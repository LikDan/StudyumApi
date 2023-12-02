package utils

import (
	"context"
	auth "studyum/internal/auth/entities"
)

type ICRUDController[T any, DTO any, ID any] interface {
	GetByID(ctx context.Context, id ID) (T, error)
	Add(ctx context.Context, value DTO) (T, error)
	Update(ctx context.Context, id ID, value DTO) (T, error)
	DeleteByID(ctx context.Context, id ID) error
}

type ICRUDControllerWithUser[T any, DTO any, ID any] interface {
	GetByID(ctx context.Context, user auth.User, id ID) (T, error)
	Add(ctx context.Context, user auth.User, value DTO) (T, error)
	Update(ctx context.Context, user auth.User, id ID, value DTO) (T, error)
	DeleteByID(ctx context.Context, user auth.User, id ID) error
}

type ICRUDRepositoryWithStudyPlaceID[T any, ID any] interface {
	GetByID(ctx context.Context, studyPlaceID, id ID) (T, error)
	Add(ctx context.Context, value T) error
	Update(ctx context.Context, studyPlaceID ID, value T) error
	DeleteByID(ctx context.Context, studyPlaceID, id ID) error
}
