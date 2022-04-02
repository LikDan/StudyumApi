package user

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	h "studyium/src/api"
	"studyium/src/db"
)

func GetUserViaGoogle(ctx *gin.Context, user *User) error {
	token, err := ctx.Cookie("authToken")

	if h.CheckError(err, h.UNDEFINED) {
		return errors.New("not authorized")
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
	if h.CheckError(err, h.UNDEFINED) || response.StatusCode != 200 {
		return errors.New("not authorized")
	}

	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if h.CheckError(err, h.WARNING) {
		return errors.New("bad callback")
	}

	var googleUser Google
	err = json.Unmarshal(content, &googleUser)
	if h.CheckError(err, h.WARNING) {
		return errors.New("bad callback")
	}

	_, err = db.UsersCollection.InsertOne(nil, googleUser)
	if err != nil {
		_, err = db.UsersCollection.UpdateOne(nil, bson.M{"_id": googleUser.Id}, bson.M{"$set": googleUser})
		if h.CheckError(err, h.WARNING) {
			return err
		}
	}

	err = db.UsersCollection.FindOne(nil, bson.M{"_id": googleUser.Id}).Decode(&user)
	if h.CheckError(err, h.WARNING) {
		return err
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

func logout(ctx *gin.Context) {
	ctx.SetCookie("authToken", "", -1, "", "", false, false)

	h.Message(ctx, 200, "successful")
}

func updateUser(ctx *gin.Context) {
	var user User
	if err := GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	if user.Applied && !h.SliceContains(user.Permissions, "editInfo") {
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

	user.Applied = false

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

type User struct {
	Id            string   `json:"id" bson:"_id"`
	Email         string   `json:"email" bson:"email"`
	VerifiedEmail bool     `json:"verifiedEmail" bson:"verifiedEmail"`
	Login         string   `json:"login" bson:"login"`
	Name          string   `json:"name" bson:"name"`
	PictureUrl    string   `json:"picture" bson:"picture"`
	Type          string   `json:"type" bson:"type"`
	TypeName      string   `json:"typeName" bson:"typeName"`
	StudyPlaceId  int      `json:"studyPlaceId" bson:"studyPlaceId"`
	Permissions   []string `json:"permissions" bson:"permissions"`
	Applied       bool     `json:"applied" bson:"applied"`
}

func BuildRequests(api *gin.RouterGroup) {
	api.GET("/auth", authorization)
	api.GET("/logout", logout)
	api.GET("/callback", callbackHandler)

	api.PUT("", updateUser)

	api.GET("", getUser)

}
