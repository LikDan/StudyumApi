package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	h "studyum/src/api"
	"studyum/src/db"
	"time"
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

func getUser(ctx *gin.Context) {
	var user User
	if err := GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	ctx.JSON(200, user)
}

func createUser(ctx *gin.Context) {
	var userData CreateUserData
	if err := ctx.BindJSON(&userData); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	if !h.CheckNotEmpty(userData.Email, userData.Login, userData.Name, userData.TypeName, userData.Type, userData.Password) || userData.PasswordRepeat != userData.Password || len(userData.Password) < 8 {
		h.ErrorMessage(ctx, "Provide valid params")
		return
	}

	user := User{
		Id:            primitive.NewObjectID(),
		Token:         h.GenerateSecureToken(),
		Password:      h.Hash(userData.Password),
		Email:         userData.Email,
		VerifiedEmail: false,
		Login:         userData.Login,
		Name:          userData.Name,
		PictureUrl:    "https://i.stack.imgur.com/l60Hf.png",
		Type:          userData.Type,
		TypeName:      userData.TypeName,
		StudyPlaceId:  userData.StudyPlaceId,
		Permissions:   nil,
		Accepted:      false,
		Blocked:       false,
	}

	if _, err := db.UsersCollection.InsertOne(nil, user); h.CheckAndMessage(ctx, 418, err, h.WARNING) {
		return
	}

	ctx.JSON(200, user)
}

func logout(ctx *gin.Context) {
	ctx.SetCookie("authToken", "", -1, "", "", false, false)

	h.Message(ctx, 200, "successful")
}

func login(ctx *gin.Context) {
	var data = struct {
		Email    string `json:"email" bson:"email"`
		Password string `json:"password" bson:"password"`
	}{}

	if err := ctx.BindJSON(&data); h.CheckAndMessage(ctx, 418, err, h.WARNING) {
		return
	}

	data.Password = h.Hash(data.Password)
	var user User
	if err := db.UsersCollection.FindOne(nil, data).Decode(&user); h.CheckAndMessage(ctx, 418, err, h.WARNING) {
		return
	}

	if user.Token == "" {
		user.Token = h.GenerateSecureToken()
		if _, err := db.UsersCollection.UpdateOne(ctx, bson.M{"email": user.Email}, bson.M{"$set": bson.M{"token": user.Token}}); err != nil {
			h.ErrorMessage(ctx, "cannot update user")
			return
		}
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:    "authToken",
		Value:   user.Token,
		Path:    "/",
		Expires: time.Now().AddDate(1, 0, 0),
	})

	ctx.JSON(200, user)
}

func updateUser(ctx *gin.Context) {
	var user User
	if err := GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	if user.Accepted && !h.SliceContains(user.Permissions, "editInfo") {
		h.ErrorMessage(ctx, "You don't have permission to edit information")
		return
	}

	var userUpdate User
	if err := ctx.Bind(&userUpdate); h.CheckAndMessage(ctx, 500, err, h.WARNING) {
		return
	}

	user.StudyPlaceId = userUpdate.StudyPlaceId
	user.Type = userUpdate.Type
	user.TypeName = userUpdate.TypeName
	user.Name = userUpdate.Name

	user.Accepted = false
	user.Blocked = false

	_, err := db.UsersCollection.UpdateOne(nil, bson.M{"_id": user.Id}, bson.M{"$set": user})
	if h.CheckAndMessage(ctx, 500, err, h.WARNING) {
		return
	}

	err = db.UsersCollection.FindOne(nil, bson.M{"_id": user.Id}).Decode(&user)
	if h.CheckAndMessage(ctx, 500, err, h.WARNING) {
		return
	}

	ctx.JSON(200, user)
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

func revokeToken(ctx *gin.Context) {
	token, err := ctx.Cookie("authToken")
	if h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	_, err = db.UsersCollection.UpdateOne(nil, bson.M{"token": token}, bson.M{"$set": bson.M{"token": ""}})
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
	api.GET("/auth", authorization)
	api.GET("/logout", logout)
	api.GET("/callback", callbackHandler)

	api.GET("", getUser)
	api.PUT("", updateUser)
	api.POST("", createUser)

	api.PUT("/login", login)

	api.PUT("/revoke", revokeToken)
	api.PUT("/token", putToken)

	api.GET("/toAccept", toAccept)
	api.PUT("/accept", accept)
	api.PUT("/decline", decline)
}
