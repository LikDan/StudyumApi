package controllers

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"studyum/src/models"
	"studyum/src/utils"
)

func (u *UserController) GetOAuth2ConfigByName(name string) *oauth2.Config {
	return Configs[name]
}

func (u *UserController) GetUserViaToken(ctx context.Context, token string) (models.User, *models.Error) {
	var user models.User
	if err := u.repository.GetUserViaToken(ctx, token, &user); err.Check() {
		return models.User{}, err
	}

	return user, models.EmptyError()
}

func (u *UserController) CallbackOAuth2(ctx context.Context, code string) (models.User, *models.Error) {
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err := models.BindError(err, 400, models.UNDEFINED); err.Check() {
		return models.User{}, err
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err := models.BindError(err, 400, models.UNDEFINED); err.Check() {
		return models.User{}, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	content, err := io.ReadAll(response.Body)
	if err := models.BindError(err, 400, models.UNDEFINED); err.Check() {
		return models.User{}, err
	}

	var googleUser models.OAuth2CallbackUser
	err = json.Unmarshal(content, &googleUser)
	if err := models.BindError(err, 400, models.UNDEFINED); err.Check() {
		return models.User{}, err
	}

	var user models.User

	if err = u.repository.GetUserByEmail(ctx, googleUser.Email, &user).Error; err != nil {
		if err.Error() != "mongo: no documents in result" {
			return models.User{}, models.BindError(err, 418, models.WARNING)
		}
		user = models.User{
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

		if err := u.repository.SignUp(ctx, &user); err.Check() {
			return models.User{}, err
		}
	}

	if user.Token == "" {
		user.Token = utils.GenerateSecureToken()

		if err := u.repository.UpdateUserTokenByEmail(ctx, user.Email, user.Token); err.Check() {
			return models.User{}, err
		}
	}

	return user, models.EmptyError()
}
