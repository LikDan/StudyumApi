package oauth2

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig = &oauth2.Config{
	ClientID:     "314976404425-5774o9r2j56p724ohicfegm6g4b2ch1t.apps.googleusercontent.com", //https://console.cloud.google.com/apis/dashboard
	ClientSecret: "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",
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://localhost:8080/api/user/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
}

var Configs = map[string]*oauth2.Config{
	"google": googleOAuthConfig,
}
