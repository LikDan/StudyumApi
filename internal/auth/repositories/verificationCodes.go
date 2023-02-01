package repositories

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"studyum/internal/auth/entities"
)

type VerificationCodes interface {
	Create(ctx context.Context, code entities.VerificationCode) error
	GetCodeByEmail(ctx context.Context, email string) (entities.VerificationCode, error)
	GetCodeAndDelete(ctx context.Context, code string) (entities.VerificationCode, error)
	DeleteAllByEmail(ctx context.Context, email string) error
}

type verificationCodes struct {
	codes *mongo.Collection
}

func NewVerificationCodes(codes *mongo.Collection) VerificationCodes {
	return &verificationCodes{codes: codes}
}

func (r *verificationCodes) Create(ctx context.Context, code entities.VerificationCode) error {
	_, err := r.codes.InsertOne(ctx, code)
	return err
}

func (r *verificationCodes) GetCodeByEmail(ctx context.Context, email string) (codeData entities.VerificationCode, err error) {
	err = r.codes.FindOne(ctx, bson.M{"email": email}).Decode(&codeData)
	return
}

func (r *verificationCodes) GetCodeAndDelete(ctx context.Context, code string) (codeData entities.VerificationCode, err error) {
	err = r.codes.FindOneAndDelete(ctx, bson.M{"code": code}).Decode(&codeData)
	return
}

func (r *verificationCodes) DeleteAllByEmail(ctx context.Context, email string) error {
	_, err := r.codes.DeleteMany(ctx, bson.M{"email": email})
	return err
}
