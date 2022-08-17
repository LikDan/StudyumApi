package repositories

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/dto"
	"studyum/internal/entities"
)

var NotAuthorizationError = errors.New("not authorized")

type UserRepository struct {
	*Repository
}

func NewUserRepository(repository *Repository) *UserRepository {
	return &UserRepository{
		Repository: repository,
	}
}

func (u *UserRepository) GetUserViaToken(ctx context.Context, token string, user *entities.User) error {
	if err := u.usersCollection.FindOne(ctx, bson.M{"token": token}).Decode(user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return NotAuthorizationError
		} else {
			return err
		}
	}

	return nil
}

func (u *UserRepository) SignUp(ctx context.Context, user *entities.User) error {
	user.Id = primitive.NewObjectID()
	if _, err := u.usersCollection.InsertOne(ctx, user); err != nil {
		return err
	}

	return nil
}

func (u *UserRepository) SignUpStage1(ctx context.Context, user *entities.User) error {
	if _, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": user.Token}, bson.M{"$set": user}); err != nil {
		return err
	}

	return nil
}

func (u *UserRepository) Login(ctx context.Context, data *dto.UserLoginData, user *entities.User) error {
	if err := u.usersCollection.FindOne(ctx, data).Decode(&user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return NotAuthorizationError
		} else {
			return err
		}
	}

	return nil
}

func (u *UserRepository) UpdateUser(ctx context.Context, user *entities.User) error {
	if _, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": user.Token}, bson.M{"$set": user}); err != nil {
		return err
	}

	return nil
}

func (u *UserRepository) RevokeToken(ctx context.Context, token string) error {
	if _, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": token}, bson.M{"$set": bson.M{"token": ""}}); err != nil {
		return err
	}

	return nil
}

func (u *UserRepository) UpdateToken(ctx context.Context, data dto.UserLoginData, token string) error {
	if _, err := u.usersCollection.UpdateOne(ctx, data, bson.M{"$set": bson.M{"token": token}}); err != nil {
		return err
	}

	return nil
}

func (u *UserRepository) UpdateUserTokenByEmail(ctx context.Context, email, token string) error {
	if _, err := u.usersCollection.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": bson.M{"token": token}}); err != nil {
		return err
	}

	return nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string, user *entities.User) error {
	if err := u.usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return err
	}

	return nil
}
