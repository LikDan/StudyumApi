package repositories

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/apps/entities"
)

var (
	ErrNotFound = errors.New("app not found")
)

type Apps interface {
	GetByStudyPlaceID(ctx context.Context, id primitive.ObjectID) (entities.App, error)
}

type apps struct {
	apps []entities.App
}

func NewApps(applications []entities.App) Apps {
	return &apps{apps: applications}
}

func (r *apps) GetByStudyPlaceID(ctx context.Context, id primitive.ObjectID) (entities.App, error) {
	for _, app := range r.apps {
		if app.GetStudyPlaceID(ctx) == id {
			return app, nil
		}
	}
	return nil, ErrNotFound
}
