package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/auth/entities"
	entities2 "studyum/internal/user/entities"
)

type Repository interface {
	GetUserByID(ctx context.Context, id primitive.ObjectID) (entities.User, error)

	UpdateUserByID(ctx context.Context, user entities.User) error

	CreateCode(ctx context.Context, code entities2.SignUpCode) error

	PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error

	GetAccept(ctx context.Context, studyPlaceID primitive.ObjectID) ([]entities2.AcceptUser, error)
	Accept(ctx context.Context, studyPlaceID primitive.ObjectID, userID primitive.ObjectID) error
	Block(ctx context.Context, studyPlaceID primitive.ObjectID, userID primitive.ObjectID) error

	GetDataByCode(ctx context.Context, code string) (entities2.SignUpCode, error)
	RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error

	GetUserByEmail(ctx context.Context, email string) (entities.User, error)

	SetPasswordByUserID(ctx context.Context, id primitive.ObjectID, password string) error
}

type repository struct {
	users       *mongo.Collection
	signupCodes *mongo.Collection
}

func NewUserRepository(users *mongo.Collection, signupCodes *mongo.Collection) Repository {
	return &repository{users: users, signupCodes: signupCodes}
}

func (u *repository) GetUserByID(ctx context.Context, id primitive.ObjectID) (user entities.User, err error) {
	err = u.users.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return
}

func (u *repository) UpdateUserByID(ctx context.Context, user entities.User) error {
	_, err := u.users.UpdateByID(ctx, user.Id, bson.M{"$set": user})
	return err
}

func (u *repository) PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error {
	_, err := u.users.UpdateByID(ctx, id, bson.M{"$set": bson.M{"firebaseToken": firebaseToken}})
	return err
}

func (u *repository) CreateCode(ctx context.Context, code entities2.SignUpCode) error {
	_, err := u.signupCodes.InsertOne(ctx, code)
	return err
}

func (u *repository) GetAccept(ctx context.Context, studyPlaceID primitive.ObjectID) (users []entities2.AcceptUser, err error) {
	cursor, err := u.users.Find(ctx, bson.M{"studyPlaceID": studyPlaceID, "accepted": false})
	if err != nil {
		return
	}

	err = cursor.All(ctx, &users)
	return
}

func (u *repository) Accept(ctx context.Context, studyPlaceID primitive.ObjectID, userID primitive.ObjectID) error {
	_, err := u.users.UpdateOne(ctx, bson.M{"studyPlaceID": studyPlaceID, "_id": userID}, bson.M{"$set": bson.M{"accepted": true}})
	return err
}

func (u *repository) Block(ctx context.Context, studyPlaceID primitive.ObjectID, userID primitive.ObjectID) error {
	_, err := u.users.UpdateOne(ctx, bson.M{"studyPlaceID": studyPlaceID, "_id": userID}, bson.M{"$set": bson.M{"blocked": true}})
	return err
}

func (u *repository) GetDataByCode(ctx context.Context, code string) (data entities2.SignUpCode, err error) {
	err = u.signupCodes.FindOne(ctx, bson.M{"code": code}).Decode(&data)
	return
}

func (u *repository) RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := u.signupCodes.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (u *repository) GetUserByEmail(ctx context.Context, email string) (user entities.User, err error) {
	err = u.users.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	return
}

func (u *repository) SetPasswordByUserID(ctx context.Context, id primitive.ObjectID, password string) error {
	_, err := u.users.UpdateByID(ctx, id, bson.M{"$set": bson.M{"password": password}})
	return err
}
