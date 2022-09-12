package controllers

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
	"studyum/internal/repositories"
)

type SignUpCodesController interface {
	GetDataByCode(ctx context.Context, code string) (entities.SignUpCode, error)
	RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error
}

type signUpCodesController struct {
	repository repositories.SignUpCodesRepository
}

func NewSignUpCodesController(repository repositories.SignUpCodesRepository) SignUpCodesController {
	return &signUpCodesController{repository: repository}
}

func (s *signUpCodesController) GetDataByCode(ctx context.Context, code string) (entities.SignUpCode, error) {
	return s.repository.GetDataByCode(ctx, code)
}

func (s *signUpCodesController) RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error {
	return s.repository.RemoveCodeByID(ctx, id)
}
