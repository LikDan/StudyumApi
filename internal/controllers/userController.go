package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"studyum/internal/dto"
	"studyum/internal/entities"
	"studyum/internal/repositories"
	"studyum/pkg/hash"
	"studyum/pkg/jwt"
	"time"
)

type UserController interface {
	UpdateUser(ctx context.Context, user entities.User, data dto.EditUserDTO) (jwt.TokenPair, error)

	LoginUser(ctx context.Context, data dto.UserLoginDTO, ip string) (entities.User, jwt.TokenPair, error)
	SignUpUser(ctx context.Context, data dto.UserSignUpDTO) (entities.User, error)
	SignUpUserStage1(ctx context.Context, user entities.User, data dto.UserSignUpStage1DTO) (entities.User, error)
	SignOut(ctx context.Context, refreshToken string) error

	RevokeToken(ctx context.Context, token string) error
	TerminateSession(ctx context.Context, user entities.User, ip string) error

	CallbackOAuth2(ctx context.Context, code string) (entities.User, error)
	GetOAuth2ConfigByName(name string) *oauth2.Config

	PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error
}

type userController struct {
	repository repositories.UserRepository

	jwt jwt.JWT[entities.JWTClaims]
}

func NewUserController(jwt jwt.JWT[entities.JWTClaims], repository repositories.UserRepository) UserController {
	return &userController{repository: repository, jwt: jwt}
}

func (u *userController) SignUpUser(ctx context.Context, data dto.UserSignUpDTO) (entities.User, error) {
	password, err := hash.Hash(data.Password)
	if err != nil {
		return entities.User{}, err
	}

	user := entities.User{
		Password:      password,
		Email:         data.Email,
		VerifiedEmail: false,
		Login:         data.Login,
		Name:          data.Name,
		PictureUrl:    "https://www.shareicon.net/data/128x128/2016/07/05/791214_man_512x512.png",
	}
	user.Id, err = u.repository.SignUp(ctx, user)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (u *userController) SignUpUserStage1(ctx context.Context, user entities.User, data dto.UserSignUpStage1DTO) (entities.User, error) {
	switch data.Type {
	case "group", "teacher":
		user.Type = data.Type
		user.StudyPlaceId = data.StudyPlaceId
		user.TypeName = data.TypeName
		break
	default:
		return entities.User{}, NotValidParams
	}

	if err := u.repository.SignUpStage1(ctx, user); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (u *userController) SignOut(ctx context.Context, refreshToken string) error {
	return u.repository.DeleteSessionByRefreshToken(ctx, refreshToken)
}

func (u *userController) UpdateUser(ctx context.Context, user entities.User, data dto.EditUserDTO) (jwt.TokenPair, error) {
	if data.Password != "" && len(data.Password) > 8 {
		password, err := hash.Hash(data.Password)
		if err != nil {
			return jwt.TokenPair{}, err
		}

		user.Password = password
	}

	user.Login = data.Login
	user.Name = data.Name
	user.Email = data.Email
	if err := u.repository.UpdateUserByID(ctx, user); err != nil {
		return jwt.TokenPair{}, err
	}

	claims := entities.JWTClaims{
		ID:            user.Id,
		Login:         user.Login,
		Permissions:   user.Permissions,
		FirebaseToken: user.FirebaseToken,
	}
	pair, err := u.jwt.GeneratePair(claims)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return pair, nil
}

func (u *userController) LoginUser(ctx context.Context, data dto.UserLoginDTO, ip string) (entities.User, jwt.TokenPair, error) {
	user, err := u.repository.GetUserByEmail(ctx, data.Email)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	if !hash.CompareHashAndPassword(user.Password, data.Password) {
		return entities.User{}, jwt.TokenPair{}, NotValidParams
	}

	claims := entities.JWTClaims{
		ID:            user.Id,
		Login:         user.Login,
		Permissions:   user.Permissions,
		FirebaseToken: user.FirebaseToken,
	}
	pair, err := u.jwt.GeneratePair(claims)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	session := entities.Session{
		RefreshToken: pair.Refresh,
		IP:           ip,
		LastOnline:   time.Now(),
	}

	if err = u.repository.AddSessionByUserID(ctx, session, user.Id, len(user.Sessions)); err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	return user, pair, nil
}

func (u *userController) RevokeToken(ctx context.Context, token string) error {
	return u.repository.RevokeToken(ctx, token)
}

func (u *userController) TerminateSession(ctx context.Context, user entities.User, ip string) error {
	return u.repository.DeleteSessionByIP(ctx, user.Id, ip)
}

func (u *userController) GetOAuth2ConfigByName(name string) *oauth2.Config {
	return Configs[name]
}

func (u *userController) CallbackOAuth2(ctx context.Context, code string) (entities.User, error) {
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

	if _, err = u.repository.GetUserByEmail(ctx, googleUser.Email); err != nil {
		if !errors.Is(mongo.ErrNoDocuments, err) {
			return entities.User{}, err
		}
		user = entities.User{
			Id:            primitive.NewObjectID(),
			Email:         googleUser.Email,
			VerifiedEmail: googleUser.VerifiedEmail,
			Login:         googleUser.Name,
			Name:          googleUser.Name,
			PictureUrl:    googleUser.PictureUrl,
			Type:          "",
			TypeName:      "",
			StudyPlaceId:  primitive.NilObjectID,
			Permissions:   nil,
			Accepted:      false,
			Blocked:       false,
		}

		user.Id, err = u.repository.SignUp(ctx, user)
		if err != nil {
			return entities.User{}, err
		}
	}

	return user, nil
}

func (u *userController) PutFirebaseTokenByUserID(ctx context.Context, token primitive.ObjectID, firebaseToken string) error {
	return u.repository.PutFirebaseTokenByUserID(ctx, token, firebaseToken)
}
