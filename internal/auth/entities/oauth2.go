package entities

import "golang.org/x/oauth2"

type OAuth2Service struct {
	oauth2.Config
	DataUrl string
}

type OAuth2ServiceRaw struct {
	ClientID     string
	ClientSecret string
	Endpoint     oauth2.Endpoint
	RedirectURL  string
	Scopes       []string
	DataUrl      string
}

func (o *OAuth2ServiceRaw) Get() OAuth2Service {
	return OAuth2Service{
		Config: oauth2.Config{
			ClientID:     o.ClientID,
			ClientSecret: o.ClientSecret,
			Endpoint:     o.Endpoint,
			RedirectURL:  o.RedirectURL,
			Scopes:       o.Scopes,
		},
		DataUrl: o.DataUrl,
	}
}

type OAuth2CallbackUser struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	PictureUrl    string `json:"picture"`
}
