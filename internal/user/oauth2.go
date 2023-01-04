package user

import "golang.org/x/oauth2"

type OAuth2 struct {
	oauth2.Config
	DataUrl string
}
