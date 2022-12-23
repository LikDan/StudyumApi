package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
)

type SignUpCodesRepository interface {
	GetDataByCode(ctx context.Context, code string) (entities.SignUpCode, error)
	RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error
}

type signUpCodesRepository struct {
	*Repository
}

func NewSignUpCodesRepository(codesRepository *Repository) SignUpCodesRepository {
	return &signUpCodesRepository{Repository: codesRepository}
}

func (s *signUpCodesRepository) GetDataByCode(ctx context.Context, code string) (data entities.SignUpCode, err error) {
	err = s.signUpCodesCollection.FindOne(ctx, bson.M{"code": code}).Decode(&data)
	return
}

func (s *signUpCodesRepository) RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := s.signUpCodesCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
