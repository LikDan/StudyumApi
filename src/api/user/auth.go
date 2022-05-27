package user

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	h "studyum/src/api"
	"studyum/src/db"
	"time"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     "314976404425-5774o9r2j56p724ohicfegm6g4b2ch1t.apps.googleusercontent.com", //https://console.cloud.google.com/apis/dashboard
	ClientSecret: "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",
	Endpoint:     google.Endpoint,
	RedirectURL:  "https://studyum-api.herokuapp.com/api/user/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
}

func authorization(ctx *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(ctx.Query("redirect"))
	ctx.Redirect(307, url)
}

func callbackHandler(ctx *gin.Context) {
	token, err := googleOauthConfig.Exchange(context.Background(), ctx.Request.FormValue("code"))
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, "bad callback")
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if h.CheckAndMessage(nil, 418, err, h.UNDEFINED) || response.StatusCode != 200 {
		return
	}

	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if h.CheckAndMessage(nil, 418, err, h.WARNING) {
		return
	}

	var googleUser Google
	err = json.Unmarshal(content, &googleUser)
	if h.CheckAndMessage(nil, 418, err, h.WARNING) {
		return
	}

	var user User
	if err = db.UsersCollection.FindOne(ctx, bson.M{"email": googleUser.Email}).Decode(&user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			user = User{
				Id:            primitive.NewObjectID(),
				Token:         h.GenerateSecureToken(),
				Email:         googleUser.Email,
				VerifiedEmail: googleUser.VerifiedEmail,
				Login:         googleUser.Name,
				Name:          "",
				PictureUrl:    googleUser.PictureUrl,
				Type:          "",
				TypeName:      "",
				StudyPlaceId:  -1,
				Permissions:   nil,
				Accepted:      false,
				Blocked:       false,
			}
			if _, err = db.UsersCollection.InsertOne(ctx, user); err != nil {
				h.ErrorMessage(ctx, "cannot create user")
				return
			}
		} else {
			h.ErrorMessage(ctx, "cannot find user")
			return
		}
	}

	if user.Token == "" {
		user.Token = h.GenerateSecureToken()
		if _, err = db.UsersCollection.UpdateOne(ctx, bson.M{"email": user.Email}, bson.M{"$set": bson.M{"token": user.Token}}); err != nil {
			h.ErrorMessage(ctx, "cannot update user")
			return
		}
	}

	ctx.Redirect(307, ctx.Request.FormValue("state")+"?token="+user.Token)
}

func putToken(ctx *gin.Context) {
	bytes, _ := ctx.GetRawData()
	token := string(bytes)

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:    "authToken",
		Value:   token,
		Path:    "/",
		Expires: time.Now().AddDate(1, 0, 0),
	})
}

type Google struct {
	Id            string `json:"id" bson:"_id"`
	Email         string `json:"email" bson:"email"`
	VerifiedEmail bool   `json:"verified_email" bson:"verifiedEmail"`
	Name          string `json:"name" bson:"login"`
	PictureUrl    string `json:"picture" bson:"picture"`
}
