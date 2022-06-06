package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	h "studyum/src/api"
	"studyum/src/models"
)

func GetUserViaToken(token string, user *models.User) *models.Error {
	if err := UsersCollection.FindOne(nil, bson.M{"token": token}).Decode(&user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return models.BindErrorStr("not authorized", 401, h.UNDEFINED)
		} else {
			return models.BindError(err, 400, h.WARNING)
		}
	}

	return models.EmptyError()
}

func SignUp(user *models.User) *models.Error {
	user.Id = primitive.NewObjectID()
	if _, err := UsersCollection.InsertOne(nil, user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return models.BindErrorStr("not authorized", 401, h.UNDEFINED)
		} else {
			return models.BindError(err, 400, h.WARNING)
		}
	}

	return models.EmptyError()
}

func SignUpStage1(user *models.User) *models.Error {
	if _, err := UsersCollection.UpdateOne(nil, bson.M{"token": user.Token}, bson.M{"$set": user}); err != nil {
		return models.BindError(err, 400, h.WARNING)
	}

	return models.EmptyError()
}

func Login(data *models.UserLoginData, user *models.User) *models.Error {
	if err := UsersCollection.FindOne(nil, data).Decode(&user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return models.BindErrorStr("not authorized", 401, h.UNDEFINED)
		} else {
			return models.BindError(err, 400, h.WARNING)
		}
	}

	return models.EmptyError()
}

func UpdateUser(user *models.User) *models.Error {
	if _, err := UsersCollection.UpdateOne(nil, bson.M{"token": user.Token}, bson.M{"$set": user}); err != nil {
		return models.BindError(err, 418, h.UNDEFINED)
	}

	return models.EmptyError()
}

func RevokeToken(token string) *models.Error {
	if _, err := UsersCollection.UpdateOne(nil, bson.M{"token": token}, bson.M{"$set": bson.M{"token": ""}}); err != nil {
		return models.BindError(err, 418, h.UNDEFINED)
	}

	return models.EmptyError()
}
