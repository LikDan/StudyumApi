package controllers

import (
	"context"
	uEntities "studyum/internal/auth/entities" //todo move to shared
	"studyum/internal/user/dto"
	"studyum/internal/user/entities"
	"studyum/internal/user/repositories"
)

type PreferencesController interface {
	GetPreferences(ctx context.Context, user uEntities.User) (entities.Preferences, error)
	SavePreferences(ctx context.Context, user uEntities.User, preferences dto.Preferences) error
}

type preferencesController struct {
	repository repositories.PreferencesRepository
}

func NewPreferencesController(repository repositories.PreferencesRepository) PreferencesController {
	return &preferencesController{repository: repository}
}

func (p *preferencesController) GetPreferences(ctx context.Context, user uEntities.User) (entities.Preferences, error) {
	return p.repository.GetPreferences(ctx, user.Id)
}

func (p *preferencesController) SavePreferences(ctx context.Context, user uEntities.User, preferencesDTO dto.Preferences) error {
	preferences := entities.Preferences{
		Theme:    preferencesDTO.Theme,
		Language: preferencesDTO.Language,
		Timezone: preferencesDTO.Timezone,
	}
	return p.repository.SavePreferences(ctx, user.Id, preferences)
}
