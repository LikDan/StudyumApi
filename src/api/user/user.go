package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	h "studyum/src/api"
	"studyum/src/db"
)

func GetUserViaGoogle(ctx *gin.Context, user *User) error {
	token, err := ctx.Cookie("authToken")
	if h.CheckError(err, h.UNDEFINED) {
		return errors.New("not authorized")
	}

	err = db.UsersCollection.FindOne(nil, bson.M{"token": token}).Decode(&user)
	if h.CheckError(err, h.WARNING) {
		if err.Error() == "mongo: no documents in result" {
			return errors.New("not authorized")
		} else {
			return err
		}
	}

	return nil
}

func toAccept(ctx *gin.Context) {
	var user User
	if err := GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	if !h.SliceContains(user.Permissions, "acceptUsers") {
		h.Message(ctx, 403, "You don't have permission to accept users")
		return
	}

	find, err := db.UsersCollection.Find(nil, bson.M{"studyPlaceId": user.StudyPlaceId, "accepted": false, "blocked": false})
	if h.CheckAndMessage(ctx, 500, err, h.WARNING) {
		return
	}

	var users []User
	err = find.All(nil, &users)

	ctx.JSON(200, users)
}

func accept(ctx *gin.Context) {
	var user User
	if err := GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	if !h.SliceContains(user.Permissions, "acceptUsers") {
		h.Message(ctx, 403, "You don't have permission to accept users")
		return
	}

	var userId string
	if err := ctx.Bind(&userId); h.CheckAndMessage(ctx, 500, err, h.WARNING) {
		return
	}

	_, err := db.UsersCollection.UpdateOne(nil, bson.M{"_id": userId}, bson.M{"$set": bson.M{"accepted": true}})
	if h.CheckAndMessage(ctx, 500, err, h.WARNING) {
		return
	}

	h.Message(ctx, 200, "successful")
}

func decline(ctx *gin.Context) {
	var user User
	if err := GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	if !h.SliceContains(user.Permissions, "acceptUsers") {
		h.Message(ctx, 403, "You don't have permission to accept users")
		return
	}

	var userId string
	if err := ctx.Bind(&userId); h.CheckAndMessage(ctx, 500, err, h.WARNING) {
		return
	}

	_, err := db.UsersCollection.UpdateOne(nil, bson.M{"_id": userId}, bson.M{"$set": bson.M{"blocked": true}})
	if h.CheckAndMessage(ctx, 500, err, h.WARNING) {
		return
	}

	h.Message(ctx, 200, "successful")
}

type User struct {
	Id            primitive.ObjectID `json:"id" bson:"_id"`
	Token         string             `json:"-" bson:"token"`
	Password      string             `json:"password" bson:"password"`
	Email         string             `json:"email" bson:"email"`
	VerifiedEmail bool               `json:"verifiedEmail" bson:"verifiedEmail"`
	Login         string             `json:"login" bson:"login"`
	Name          string             `json:"name" bson:"name"`
	PictureUrl    string             `json:"picture" bson:"picture"`
	Type          string             `json:"type" bson:"type"`
	TypeName      string             `json:"typeName" bson:"typeName"`
	StudyPlaceId  int                `json:"studyPlaceId" bson:"studyPlaceId"`
	Permissions   []string           `json:"permissions" bson:"permissions"`
	Accepted      bool               `json:"accepted" bson:"accepted"`
	Blocked       bool               `json:"blocked" bson:"blocked"`
}

type CreateUserData struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Login          string `json:"login"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	TypeName       string `json:"typeName"`
	StudyPlaceId   int    `json:"studyPlaceId"`
	Password       string `json:"password"`
	PasswordRepeat string `json:"passwordRepeat"`
}

func BuildRequests(api *gin.RouterGroup) {
	api.GET("/toAccept", toAccept)
	api.PUT("/accept", accept)
	api.PUT("/decline", decline)
}
