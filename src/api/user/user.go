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

func GetUserFromDbViaCookies(ctx *gin.Context) (*User, error) {
	login, loginErr := ctx.Cookie("info")
	token, tokenErr := ctx.Cookie("token")

	if h.CheckError(loginErr, h.UNDEFINED) || h.CheckError(tokenErr, h.UNDEFINED) {
		return nil, errors.New("not authorized")
	}

	var user User

	userResult := db.UsersCollection.FindOne(nil, bson.M{"info": login, "token": token})
	err := userResult.Decode(&user)
	if h.CheckError(err, h.UNDEFINED) {
		return nil, errors.New("not authorized")
	}

	return &user, nil
}

func info(ctx *gin.Context) {
	token, err := ctx.Cookie("authToken")

	if err != nil || token == "" {
		h.ErrorMessage(ctx, "not authorized")

		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
	if h.CheckError(err, h.UNDEFINED) || response.StatusCode != 200 {
		h.ErrorMessage(ctx, "not authorized")
		return
	}

	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, "bad callback")
		return
	}

	var googleUser Google
	err = json.Unmarshal(content, &googleUser)
	if err != nil {
		return
	}

	_, err = db.UsersCollection.InsertOne(nil, googleUser)
	if err != nil {
		_, err = db.UsersCollection.UpdateOne(nil, bson.M{"_id": googleUser.Id}, bson.M{"$set": googleUser})
		if h.CheckError(err, h.WARNING) {
			h.ErrorMessage(ctx, err.Error())
			return
		}
	}

	var user User
	err = db.UsersCollection.FindOne(nil, bson.M{"_id": googleUser.Id}).Decode(&user)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	ctx.JSON(200, user)
}

func logout(ctx *gin.Context) {
	ctx.SetCookie("authToken", "", -1, "", "", false, false)

	h.Message(ctx, "message", "successful", 200)
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

	api.GET("", info)

}
