package controllers

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"os"
	"studyum/internal/entities"
)

var googleOAuthConfig = &entities.OAuth2{
	Config: oauth2.Config{
		ClientID:     "923545242743-r22djbfqvaugug2c6o3tntdgh3kn86ah.apps.googleusercontent.com", //https://console.cloud.google.com/apis/dashboard
		ClientSecret: os.Getenv("GOOGLE_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  os.Getenv("OAUTH2_CALLBACK") + "/api/user/oauth2/callback/google",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	},
	DataUrl: "https://www.googleapis.com/oauth2/v2/userinfo?access_token=",
}

var Configs = map[string]*entities.OAuth2{
	"google": googleOAuthConfig,
}
