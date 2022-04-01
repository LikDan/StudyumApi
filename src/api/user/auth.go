package user

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	h "studyium/src/api"
	"studyium/src/db"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     "314976404425-5774o9r2j56p724ohicfegm6g4b2ch1t.apps.googleusercontent.com", //https://console.cloud.google.com/apis/dashboard
	ClientSecret: "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",
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://localhost:8080/api/user/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
}

var state = "2c603d83c97ce606120435bc55416495"

func Authorization(ctx *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(state)
	ctx.Redirect(307, url)
}

func CallbackHandler(ctx *gin.Context) {
	if ctx.Request.FormValue("state") != state {
		h.ErrorMessage(ctx, "bad callback")
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), ctx.Request.FormValue("code"))
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, "bad callback")
		return
	}

	ctx.SetCookie("authToken", token.AccessToken, 0, "", "", false, false)

	ctx.Redirect(307, "/api/user/login")
}

func Login(ctx *gin.Context) {
	token, err := ctx.Cookie("authToken")
	if err != nil || token == "" {
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, "bad callback")
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

	var user U
	err = db.UsersCollection.FindOne(nil, bson.M{"_id": googleUser.Id}).Decode(&user)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	ctx.JSON(200, user)
}

type Google struct {
	Id            string `json:"id" bson:"_id"`
	Email         string `json:"email" bson:"email"`
	VerifiedEmail bool   `json:"verified_email" bson:"verifiedEmail"`
	Name          string `json:"name" bson:"name"`
	PictureUrl    string `json:"picture" bson:"picture"`
}

type U struct {
	Id            string `json:"id" bson:"_id"`
	Email         string `json:"email" bson:"email"`
	VerifiedEmail bool   `json:"verified_email" bson:"verifiedEmail"`
	Name          string `json:"name" bson:"name"`
	PictureUrl    string `json:"picture" bson:"picture"`
}
