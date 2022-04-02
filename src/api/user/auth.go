package user

import (
	"context"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
	h "studyium/src/api"
	"time"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     "314976404425-5774o9r2j56p724ohicfegm6g4b2ch1t.apps.googleusercontent.com", //https://console.cloud.google.com/apis/dashboard
	ClientSecret: "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",
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://localhost:8080/api/user/callback",
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

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:    "authToken",
		Value:   token.AccessToken,
		Path:    "/",
		Expires: time.Now().AddDate(1, 0, 0),
	})

	ctx.Redirect(307, ctx.Request.FormValue("state"))
}

type Google struct {
	Id            string `json:"id" bson:"_id"`
	Email         string `json:"email" bson:"email"`
	VerifiedEmail bool   `json:"verified_email" bson:"verifiedEmail"`
	Name          string `json:"name" bson:"login"`
	PictureUrl    string `json:"picture" bson:"picture"`
}
