package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
)

type UserRepository struct {
	*Repository
}

func NewUserRepository(repository *Repository) *UserRepository {
	return &UserRepository{
		Repository: repository,
	}
}

func (u *UserRepository) GetUserViaToken(ctx context.Context, token string, user *models.User) *models.Error {
	if err := u.usersCollection.FindOne(ctx, bson.M{"token": token}).Decode(user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return models.BindErrorStr("not authorized", 401, models.UNDEFINED)
		} else {
			return models.BindError(err, 400, models.WARNING)
		}
	}

	return models.EmptyError()
}

func (u *UserRepository) SignUp(ctx context.Context, user *models.User) *models.Error {
	user.Id = primitive.NewObjectID()
	if _, err := u.usersCollection.InsertOne(ctx, user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return models.BindErrorStr("not authorized", 401, models.UNDEFINED)
		} else {
			return models.BindError(err, 400, models.WARNING)
		}
	}

	return models.EmptyError()
}

func (u *UserRepository) SignUpStage1(ctx context.Context, user *models.User) *models.Error {
	if _, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": user.Token}, bson.M{"$set": user}); err != nil {
		return models.BindError(err, 400, models.WARNING)
	}

	return models.EmptyError()
}

func (u *UserRepository) Login(ctx context.Context, data *models.UserLoginData, user *models.User) *models.Error {
	if err := u.usersCollection.FindOne(ctx, data).Decode(&user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return models.BindErrorStr("not authorized", 401, models.UNDEFINED)
		} else {
			return models.BindError(err, 400, models.WARNING)
		}
	}

	return models.EmptyError()
}

func (u *UserRepository) UpdateUser(ctx context.Context, user *models.User) *models.Error {
	if _, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": user.Token}, bson.M{"$set": user}); err != nil {
		return models.BindError(err, 418, models.UNDEFINED)
	}

	return models.EmptyError()
}

func (u *UserRepository) RevokeToken(ctx context.Context, token string) *models.Error {
	if _, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": token}, bson.M{"$set": bson.M{"token": ""}}); err != nil {
		return models.BindError(err, 418, models.UNDEFINED)
	}

	return models.EmptyError()
}

func (u *UserRepository) UpdateToken(ctx context.Context, data models.UserLoginData, token string) *models.Error {
	if _, err := u.usersCollection.UpdateOne(ctx, data, bson.M{"$set": bson.M{"token": token}}); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func (u *UserRepository) UpdateUserTokenByEmail(ctx context.Context, email, token string) *models.Error {
	if _, err := u.usersCollection.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": bson.M{"token": token}}); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string, user *models.User) *models.Error {
	if err := u.usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return models.BindError(err, 500, models.WARNING)
	}

	return models.EmptyError()
}
