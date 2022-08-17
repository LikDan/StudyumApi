package controllers

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"studyum/internal/entities"
	"studyum/internal/utils"
)

func (u *UserController) GetOAuth2ConfigByName(name string) *oauth2.Config {
	return Configs[name]
}

func (u *UserController) GetUserViaToken(ctx context.Context, token string) (entities.User, error) {
	var user entities.User
	if err := u.repository.GetUserViaToken(ctx, token, &user); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (u *UserController) CallbackOAuth2(ctx context.Context, code string) (entities.User, error) {
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return entities.User{}, err
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return entities.User{}, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return entities.User{}, err
	}

	var googleUser entities.OAuth2CallbackUser
	err = json.Unmarshal(content, &googleUser)
	if err != nil {
		return entities.User{}, err
	}

	var user entities.User

	if err = u.repository.GetUserByEmail(ctx, googleUser.Email, &user); err != nil {
		if err.Error() != "mongo: no documents in result" {
			return entities.User{}, err
		}
		user = entities.User{
			Id:            primitive.NewObjectID(),
			Token:         utils.GenerateSecureToken(),
			Email:         googleUser.Email,
			VerifiedEmail: googleUser.VerifiedEmail,
			Login:         googleUser.Name,
			Name:          googleUser.Name,
			PictureUrl:    googleUser.PictureUrl,
			Type:          "",
			TypeName:      "",
			StudyPlaceId:  0,
			Permissions:   nil,
			Accepted:      false,
			Blocked:       false,
		}

		if err := u.repository.SignUp(ctx, &user); err != nil {
			return entities.User{}, err
		}
	}

	if user.Token == "" {
		user.Token = utils.GenerateSecureToken()

		if err = u.repository.UpdateUserTokenByEmail(ctx, user.Email, user.Token); err != nil {
			return entities.User{}, err
		}
	}

	return user, nil
}
