package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
)

type UserRepository interface {
	GetUserViaToken(ctx context.Context, token string, permissions ...string) (entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (entities.User, error)

	SignUp(ctx context.Context, user entities.User) (primitive.ObjectID, error)
	SignUpStage1(ctx context.Context, user entities.User) error

	Login(ctx context.Context, email string, password string) (entities.User, error)

	UpdateUser(ctx context.Context, user entities.User) error

	RevokeToken(ctx context.Context, token string) error
	UpdateToken(ctx context.Context, id primitive.ObjectID, token string) error
	UpdateUserTokenByEmail(ctx context.Context, email, token string) error
}

type userRepository struct {
	*Repository
}

func NewUserRepository(repository *Repository) UserRepository {
	return &userRepository{Repository: repository}
}

func (u *userRepository) GetUserViaToken(ctx context.Context, token string, permissions ...string) (entities.User, error) {
	var user entities.User

	var filter bson.M
	if len(permissions) == 0 {
		filter = bson.M{"token": token}
	} else {
		filter = bson.M{"token": token, "permissions": bson.M{"$all": permissions}}
	}
	err := u.usersCollection.FindOne(ctx, filter).Decode(user)
	return user, err
}

func (u *userRepository) SignUp(ctx context.Context, user entities.User) (primitive.ObjectID, error) {
	user.Id = primitive.NewObjectID()
	if _, err := u.usersCollection.InsertOne(ctx, user); err != nil {
		return primitive.NilObjectID, err
	}

	return user.Id, nil
}

func (u *userRepository) SignUpStage1(ctx context.Context, user entities.User) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": user.Token}, bson.M{"$set": user})
	return err
}

func (u *userRepository) Login(ctx context.Context, email string, password string) (entities.User, error) {
	var user entities.User
	err := u.usersCollection.FindOne(ctx, bson.M{"email": email, "password": password}).Decode(&user)

	return user, err
}

func (u *userRepository) UpdateUser(ctx context.Context, user entities.User) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": user.Token}, bson.M{"$set": user})
	return err
}

func (u *userRepository) RevokeToken(ctx context.Context, token string) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": token}, bson.M{"$set": bson.M{"token": ""}})
	return err
}

func (u *userRepository) UpdateToken(ctx context.Context, id primitive.ObjectID, token string) error {
	_, err := u.usersCollection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"token": token}})
	return err
}

func (u *userRepository) UpdateUserTokenByEmail(ctx context.Context, email, token string) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": bson.M{"token": token}})
	return err
}

func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	var user entities.User
	err := u.usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	return user, err
}
