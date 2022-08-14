package db

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
)

func GetUserViaToken(token string, user *models.User) *models.Error {
	if err := usersCollection.FindOne(nil, bson.M{"token": token}).Decode(user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return models.BindErrorStr("not authorized", 401, models.UNDEFINED)
		} else {
			return models.BindError(err, 400, models.WARNING)
		}
	}

	return models.EmptyError()
}

func SignUp(user *models.User) *models.Error {
	user.Id = primitive.NewObjectID()
	if _, err := usersCollection.InsertOne(nil, user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return models.BindErrorStr("not authorized", 401, models.UNDEFINED)
		} else {
			return models.BindError(err, 400, models.WARNING)
		}
	}

	return models.EmptyError()
}

func SignUpStage1(user *models.User) *models.Error {
	if _, err := usersCollection.UpdateOne(nil, bson.M{"token": user.Token}, bson.M{"$set": user}); err != nil {
		return models.BindError(err, 400, models.WARNING)
	}

	return models.EmptyError()
}

func Login(data *models.UserLoginData, user *models.User) *models.Error {
	if err := usersCollection.FindOne(nil, data).Decode(&user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return models.BindErrorStr("not authorized", 401, models.UNDEFINED)
		} else {
			return models.BindError(err, 400, models.WARNING)
		}
	}

	return models.EmptyError()
}

func UpdateUser(user *models.User) *models.Error {
	if _, err := usersCollection.UpdateOne(nil, bson.M{"token": user.Token}, bson.M{"$set": user}); err != nil {
		return models.BindError(err, 418, models.UNDEFINED)
	}

	return models.EmptyError()
}

func RevokeToken(token string) *models.Error {
	if _, err := usersCollection.UpdateOne(nil, bson.M{"token": token}, bson.M{"$set": bson.M{"token": ""}}); err != nil {
		return models.BindError(err, 418, models.UNDEFINED)
	}

	return models.EmptyError()
}

func UpdateToken(ctx context.Context, data models.UserLoginData, token string) *models.Error {
	if _, err := usersCollection.UpdateOne(ctx, data, bson.M{"$set": bson.M{"token": token}}); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func UpdateUserTokenByEmail(ctx context.Context, email, token string) *models.Error {
	if _, err := usersCollection.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": bson.M{"token": token}}); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func GetUserByEmail(ctx context.Context, email string, user *models.User) *models.Error {
	if err := usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return models.BindError(err, 500, models.WARNING)
	}

	return models.EmptyError()
}
