package controllers

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"studyum/internal/dto"
	"studyum/internal/entities"
	parser "studyum/internal/parser/handler"
	"studyum/internal/repositories"
	"studyum/pkg/encryption"
	"studyum/pkg/hash"
	"studyum/pkg/jwt"
	"time"
)

type UserController interface {
	UpdateUser(ctx context.Context, user entities.User, data dto.EditUserDTO) (entities.User, jwt.TokenPair, error)

	LoginUser(ctx context.Context, data dto.UserLoginDTO, ip string) (entities.User, jwt.TokenPair, error)
	SignUpUser(ctx context.Context, data dto.UserSignUpDTO, ip string) (entities.User, jwt.TokenPair, error)
	SignUpUserStage1(ctx context.Context, user entities.User, data dto.UserSignUpStage1DTO) (entities.User, error)
	SignUpUserWithCode(ctx context.Context, ip string, data dto.UserSignUpWithCodeDTO) (entities.User, jwt.TokenPair, error)
	SignOut(ctx context.Context, refreshToken string) error

	CreateCode(ctx context.Context, user entities.User, data dto.UserCreateCodeDTO) (entities.SignUpCode, error)

	RevokeToken(ctx context.Context, token string) error
	TerminateSession(ctx context.Context, user entities.User, ip string) error

	CallbackOAuth2(ctx context.Context, configName string, code string) (jwt.TokenPair, error)
	GetOAuth2ConfigByName(name string) *entities.OAuth2

	PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error

	GetAccept(ctx context.Context, user entities.User) ([]entities.AcceptUser, error)
	Accept(ctx context.Context, user entities.User, acceptUserID primitive.ObjectID) error
	Block(ctx context.Context, user entities.User, blockUserID primitive.ObjectID) error
}

type userController struct {
	repository            repositories.UserRepository
	signUpCodesController SignUpCodesController

	jwt     jwt.JWT[entities.JWTClaims]
	encrypt encryption.Encryption

	parser parser.Handler
}

func NewUserController(jwt jwt.JWT[entities.JWTClaims], signUpCodesController SignUpCodesController, repository repositories.UserRepository, encrypt encryption.Encryption, parser parser.Handler) UserController {
	return &userController{repository: repository, signUpCodesController: signUpCodesController, jwt: jwt, encrypt: encrypt, parser: parser}
}

func (u *userController) SignUpUser(ctx context.Context, data dto.UserSignUpDTO, ip string) (entities.User, jwt.TokenPair, error) {
	password, err := hash.Hash(data.Password)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	if exUser, _ := u.repository.GetUserByLogin(ctx, data.Login); exUser.Login != "" {
		return entities.User{}, jwt.TokenPair{}, errors.Wrap(NotValidParams, "existing login")
	}

	user := entities.User{
		Id:            primitive.NewObjectID(),
		Password:      password,
		Email:         data.Email,
		VerifiedEmail: false,
		Login:         data.Login,
		Name:          data.Name,
		PictureUrl:    "https://www.shareicon.net/data/128x128/2016/07/05/791214_man_512x512.png",
	}

	u.encrypt.Encrypt(&user)
	user.Id, err = u.repository.SignUp(ctx, user)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	u.encrypt.Decrypt(&user)

	loginData := dto.UserLoginDTO{
		Login:    data.Login,
		Password: data.Password,
	}
	return u.LoginUser(ctx, loginData, ip)
}

func (u *userController) SignUpUserStage1(ctx context.Context, user entities.User, data dto.UserSignUpStage1DTO) (entities.User, error) {
	switch data.Type {
	case "group":
		user.Type = data.Type
		user.StudyPlaceID = data.StudyPlaceId
		user.TypeName = data.TypeName
		break
	case "teacher":
		user.Type = data.Type
		user.StudyPlaceID = data.StudyPlaceId
		user.TypeName = data.TypeName
		user.Permissions = []string{"editJournal"}
		break
	default:
		return entities.User{}, NotValidParams
	}

	u.encrypt.Encrypt(&user)
	if err := u.repository.SignUpStage1(ctx, user); err != nil {
		return entities.User{}, err
	}

	u.encrypt.Decrypt(&user)
	return user, nil
}

func (u *userController) SignUpUserWithCode(ctx context.Context, ip string, data dto.UserSignUpWithCodeDTO) (entities.User, jwt.TokenPair, error) {
	codeData, err := u.signUpCodesController.GetDataByCode(ctx, data.Code)
	if err != nil {
		appCodeData, err2 := u.parser.GetSignUpDataByCode(ctx, data.Code)
		if err2 != nil {
			return entities.User{}, jwt.TokenPair{}, err
		}

		codeData = appCodeData
	}

	u.encrypt.Decrypt(&codeData)

	password, err := hash.Hash(data.Password)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	user := entities.User{
		Id:           codeData.Id,
		Password:     password,
		Email:        data.Email,
		Login:        data.Login,
		Name:         codeData.Name,
		PictureUrl:   "https://www.shareicon.net/data/128x128/2016/07/05/791214_man_512x512.png",
		Type:         codeData.Type,
		TypeName:     codeData.Typename,
		StudyPlaceID: codeData.StudyPlaceID,
		Accepted:     true,
	}

	u.encrypt.Encrypt(&user)
	user.Id, err = u.repository.SignUp(ctx, user)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	if codeData.Id != primitive.NilObjectID {
		if err = u.signUpCodesController.RemoveCodeByID(ctx, codeData.Id); err != nil {
			logrus.Warn("cannot remove code in db err: " + err.Error())
		}
	}

	loginData := dto.UserLoginDTO{
		Login:    data.Login,
		Password: data.Password,
	}

	return u.LoginUser(ctx, loginData, ip)
}

func (u *userController) SignOut(ctx context.Context, refreshToken string) error {
	return u.repository.DeleteSessionByRefreshToken(ctx, refreshToken)
}

func (u *userController) UpdateUser(ctx context.Context, user entities.User, data dto.EditUserDTO) (entities.User, jwt.TokenPair, error) {
	if data.Password != "" && len(data.Password) > 8 {
		password, err := hash.Hash(data.Password)
		if err != nil {
			return entities.User{}, jwt.TokenPair{}, err
		}

		user.Password = password
	}

	user.Login = data.Login
	user.Email = data.Email
	user.PictureUrl = data.Picture

	u.encrypt.Encrypt(&user)
	if err := u.repository.UpdateUserByID(ctx, user); err != nil {
		return entities.User{}, jwt.TokenPair{}, err
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

	u.encrypt.Decrypt(&user)
	return user, pair, nil
}

func (u *userController) LoginUser(ctx context.Context, data dto.UserLoginDTO, ip string) (entities.User, jwt.TokenPair, error) {
	user, err := u.repository.GetUserByLogin(ctx, data.Login)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	u.encrypt.Decrypt(&user)
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

func (u *userController) GetOAuth2ConfigByName(name string) *entities.OAuth2 {
	return Configs[name]
}

func (u *userController) CallbackOAuth2(ctx context.Context, configName string, code string) (jwt.TokenPair, error) {
	config := u.GetOAuth2ConfigByName(configName)

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	response, err := http.Get(config.DataUrl + token.AccessToken)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	var callbackUser entities.OAuth2CallbackUser
	if err = json.Unmarshal(content, &callbackUser); err != nil {
		return jwt.TokenPair{}, err
	}

	user, err := u.repository.GetUserByLogin(ctx, callbackUser.Email)
	if err != nil {
		if !errors.Is(mongo.ErrNoDocuments, err) {
			return jwt.TokenPair{}, err
		}
		user = entities.User{
			Id:            primitive.NewObjectID(),
			Email:         callbackUser.Email,
			VerifiedEmail: callbackUser.VerifiedEmail,
			Login:         callbackUser.Name,
			Name:          callbackUser.Name,
			PictureUrl:    callbackUser.PictureUrl,
			Type:          "",
			TypeName:      "",
			StudyPlaceID:  primitive.NilObjectID,
			Permissions:   nil,
			Accepted:      false,
			Blocked:       false,
		}

		u.encrypt.Encrypt(&user)
		user.Id, err = u.repository.SignUp(ctx, user)
		if err != nil {
			return jwt.TokenPair{}, err
		}
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

func (u *userController) PutFirebaseTokenByUserID(ctx context.Context, token primitive.ObjectID, firebaseToken string) error {
	return u.repository.PutFirebaseTokenByUserID(ctx, token, firebaseToken)
}

func (u *userController) CreateCode(ctx context.Context, user entities.User, data dto.UserCreateCodeDTO) (entities.SignUpCode, error) {
	code := entities.SignUpCode{
		Id:           primitive.NewObjectID(),
		Code:         data.Code,
		Name:         data.Name,
		StudyPlaceID: user.StudyPlaceID,
		Type:         data.Type,
		Typename:     data.TypeName,
	}

	u.encrypt.Encrypt(&code)
	if err := u.repository.CreateCode(ctx, code); err != nil {
		return entities.SignUpCode{}, nil
	}

	u.encrypt.Decrypt(&code)
	return code, nil
}

func (u *userController) GetAccept(ctx context.Context, user entities.User) ([]entities.AcceptUser, error) {
	users, err := u.repository.GetAccept(ctx, user.StudyPlaceID)
	if err != nil {
		return nil, err
	}

	u.encrypt.Decrypt(&users)
	return users, nil
}

func (u *userController) Accept(ctx context.Context, user entities.User, acceptUserID primitive.ObjectID) error {
	return u.repository.Accept(ctx, user.StudyPlaceID, acceptUserID)
}

func (u *userController) Block(ctx context.Context, user entities.User, blockUserID primitive.ObjectID) error {
	return u.repository.Block(ctx, user.StudyPlaceID, blockUserID)
}
