package db

import (
	"go.mongodb.org/mongo-driver/bson"
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
