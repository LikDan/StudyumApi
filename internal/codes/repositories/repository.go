package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/codes/entities"
)

type Repository interface {
	Create(ctx context.Context, code entities.Code) error
	GetCodeByEmail(ctx context.Context, email string) (entities.Code, error)
	GetCodeByUserID(ctx context.Context, id primitive.ObjectID) (entities.Code, error)
	GetCodeAndDelete(ctx context.Context, code string) (entities.Code, error)
	DeleteAllByEmail(ctx context.Context, email string) error
	DeleteAllByUserID(ctx context.Context, id primitive.ObjectID) error
}

type repository struct {
	codes *mongo.Collection
}

func New(codes *mongo.Collection) Repository {
	return &repository{codes: codes}
}

func (r *repository) Create(ctx context.Context, code entities.Code) error {
	_, err := r.codes.InsertOne(ctx, code)
	return err
}

func (r *repository) GetCodeByEmail(ctx context.Context, email string) (code entities.Code, err error) {
	err = r.codes.FindOne(ctx, bson.M{"email": email}).Decode(&code)
	return
}

func (r *repository) GetCodeByUserID(ctx context.Context, userID primitive.ObjectID) (code entities.Code, err error) {
	err = r.codes.FindOne(ctx, bson.M{"userID": userID}).Decode(&code)
	return
}

func (r *repository) GetCodeAndDelete(ctx context.Context, code string) (codeData entities.Code, err error) {
	err = r.codes.FindOneAndDelete(ctx, bson.M{"code": code}).Decode(&codeData)
	return
}

func (r *repository) DeleteAllByEmail(ctx context.Context, email string) error {
	_, err := r.codes.DeleteMany(ctx, bson.M{"email": email})
	return err
}

func (r *repository) DeleteAllByUserID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.codes.DeleteMany(ctx, bson.M{"userID": userID})
	return err
}
