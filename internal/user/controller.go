package user

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"studyum/internal/global"
	parser "studyum/internal/parser/handler"
	"studyum/pkg/encryption"
	"studyum/pkg/hash"
	"studyum/pkg/jwt"
	"time"
)

type Controller interface {
	UpdateUser(ctx context.Context, user global.User, data EditUserDTO) (global.User, jwt.TokenPair, error)

	LoginUser(ctx context.Context, data UserLoginDTO, ip string) (global.User, jwt.TokenPair, error)
	SignUpUser(ctx context.Context, data UserSignUpDTO, ip string) (global.User, jwt.TokenPair, error)
	SignUpUserStage1(ctx context.Context, user global.User, data UserSignUpStage1DTO) (global.User, error)
	SignUpUserWithCode(ctx context.Context, ip string, data UserSignUpWithCodeDTO) (global.User, jwt.TokenPair, error)
	SignOut(ctx context.Context, refreshToken string) error

	CreateCode(ctx context.Context, user global.User, data UserCreateCodeDTO) (SignUpCode, error)

	RevokeToken(ctx context.Context, token string) error
	TerminateSession(ctx context.Context, user global.User, ip string) error

	CallbackOAuth2(ctx context.Context, configName string, code string) (jwt.TokenPair, error)
	GetOAuth2ConfigByName(name string) *OAuth2

	PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error

	GetAccept(ctx context.Context, user global.User) ([]AcceptUser, error)
	Accept(ctx context.Context, user global.User, acceptUserID primitive.ObjectID) error
	Block(ctx context.Context, user global.User, blockUserID primitive.ObjectID) error

	GetDataByCode(ctx context.Context, code string) (SignUpCode, error)
	RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error
}

type controller struct {
	repository Repository

	jwt     jwt.JWT[global.JWTClaims]
	encrypt encryption.Encryption

	parser parser.Handler
}

func NewUserController(jwt jwt.JWT[global.JWTClaims], repository Repository, encrypt encryption.Encryption, parser parser.Handler) Controller {
	return &controller{repository: repository, jwt: jwt, encrypt: encrypt, parser: parser}
}

func (u *controller) SignUpUser(ctx context.Context, data UserSignUpDTO, ip string) (global.User, jwt.TokenPair, error) {
	password, err := hash.Hash(data.Password)
	if err != nil {
		return global.User{}, jwt.TokenPair{}, err
	}

	if exUser, _ := u.repository.GetUserByLogin(ctx, data.Login); exUser.Login != "" {
		return global.User{}, jwt.TokenPair{}, errors.Wrap(global.NotValidParams, "existing login")
	}

	user := global.User{
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
		return global.User{}, jwt.TokenPair{}, err
	}

	u.encrypt.Decrypt(&user)

	loginData := UserLoginDTO{
		Login:    data.Login,
		Password: data.Password,
	}
	return u.LoginUser(ctx, loginData, ip)
}

func (u *controller) SignUpUserStage1(ctx context.Context, user global.User, data UserSignUpStage1DTO) (global.User, error) {
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
		return global.User{}, global.NotValidParams
	}

	u.encrypt.Encrypt(&user)
	if err := u.repository.SignUpStage1(ctx, user); err != nil {
		return global.User{}, err
	}

	u.encrypt.Decrypt(&user)
	return user, nil
}

func (u *controller) SignUpUserWithCode(ctx context.Context, ip string, data UserSignUpWithCodeDTO) (global.User, jwt.TokenPair, error) {
	codeData, err := u.GetDataByCode(ctx, data.Code)
	if err != nil {
		return global.User{}, jwt.TokenPair{}, err
	}

	u.encrypt.Decrypt(&codeData)

	password, err := hash.Hash(data.Password)
	if err != nil {
		return global.User{}, jwt.TokenPair{}, err
	}

	user := global.User{
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
		return global.User{}, jwt.TokenPair{}, err
	}

	if codeData.Id != primitive.NilObjectID {
		if err = u.RemoveCodeByID(ctx, codeData.Id); err != nil {
			logrus.Warn("cannot remove code in db err: " + err.Error())
		}
	}

	loginData := UserLoginDTO{
		Login:    data.Login,
		Password: data.Password,
	}

	return u.LoginUser(ctx, loginData, ip)
}

func (u *controller) SignOut(ctx context.Context, refreshToken string) error {
	return u.repository.DeleteSessionByRefreshToken(ctx, refreshToken)
}

func (u *controller) UpdateUser(ctx context.Context, user global.User, data EditUserDTO) (global.User, jwt.TokenPair, error) {
	if data.Password != "" && len(data.Password) > 8 {
		password, err := hash.Hash(data.Password)
		if err != nil {
			return global.User{}, jwt.TokenPair{}, err
		}

		user.Password = password
	}

	user.Login = data.Login
	user.Email = data.Email
	user.PictureUrl = data.Picture

	u.encrypt.Encrypt(&user)
	if err := u.repository.UpdateUserByID(ctx, user); err != nil {
		return global.User{}, jwt.TokenPair{}, err
	}

	claims := global.JWTClaims{
		ID:            user.Id,
		Login:         user.Login,
		Permissions:   user.Permissions,
		FirebaseToken: user.FirebaseToken,
	}

	pair, err := u.jwt.GeneratePair(claims)
	if err != nil {
		return global.User{}, jwt.TokenPair{}, err
	}

	u.encrypt.Decrypt(&user)
	return user, pair, nil
}

func (u *controller) LoginUser(ctx context.Context, data UserLoginDTO, ip string) (global.User, jwt.TokenPair, error) {
	user, err := u.repository.GetUserByLogin(ctx, data.Login)
	if err != nil {
		return global.User{}, jwt.TokenPair{}, err
	}

	u.encrypt.Decrypt(&user)
	if !hash.CompareHashAndPassword(user.Password, data.Password) {
		return global.User{}, jwt.TokenPair{}, global.NotValidParams
	}

	claims := global.JWTClaims{
		ID:            user.Id,
		Login:         user.Login,
		Permissions:   user.Permissions,
		FirebaseToken: user.FirebaseToken,
	}
	pair, err := u.jwt.GeneratePair(claims)
	if err != nil {
		return global.User{}, jwt.TokenPair{}, err
	}

	session := global.Session{
		RefreshToken: pair.Refresh,
		IP:           ip,
		LastOnline:   time.Now(),
	}

	if err = u.repository.AddSessionByUserID(ctx, session, user.Id, len(user.Sessions)); err != nil {
		return global.User{}, jwt.TokenPair{}, err
	}

	return user, pair, nil
}

func (u *controller) RevokeToken(ctx context.Context, token string) error {
	return u.repository.RevokeToken(ctx, token)
}

func (u *controller) TerminateSession(ctx context.Context, user global.User, ip string) error {
	return u.repository.DeleteSessionByIP(ctx, user.Id, ip)
}

func (u *controller) GetOAuth2ConfigByName(name string) *OAuth2 {
	return Configs[name]
}

func (u *controller) CallbackOAuth2(ctx context.Context, configName string, code string) (jwt.TokenPair, error) {
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

	var callbackUser OAuth2CallbackUser
	if err = json.Unmarshal(content, &callbackUser); err != nil {
		return jwt.TokenPair{}, err
	}

	user, err := u.repository.GetUserByLogin(ctx, callbackUser.Email)
	if err != nil {
		if !errors.Is(mongo.ErrNoDocuments, err) {
			return jwt.TokenPair{}, err
		}
		user = global.User{
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

	claims := global.JWTClaims{
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

func (u *controller) PutFirebaseTokenByUserID(ctx context.Context, token primitive.ObjectID, firebaseToken string) error {
	return u.repository.PutFirebaseTokenByUserID(ctx, token, firebaseToken)
}

func (u *controller) CreateCode(ctx context.Context, user global.User, data UserCreateCodeDTO) (SignUpCode, error) {
	code := SignUpCode{
		Id:           primitive.NewObjectID(),
		Code:         data.Code,
		Name:         data.Name,
		StudyPlaceID: user.StudyPlaceID,
		Type:         data.Type,
		Typename:     data.TypeName,
	}

	u.encrypt.Encrypt(&code)
	if err := u.repository.CreateCode(ctx, code); err != nil {
		return SignUpCode{}, nil
	}

	u.encrypt.Decrypt(&code)
	return code, nil
}

func (u *controller) GetAccept(ctx context.Context, user global.User) ([]AcceptUser, error) {
	users, err := u.repository.GetAccept(ctx, user.StudyPlaceID)
	if err != nil {
		return nil, err
	}

	u.encrypt.Decrypt(&users)
	return users, nil
}

func (u *controller) Accept(ctx context.Context, user global.User, acceptUserID primitive.ObjectID) error {
	return u.repository.Accept(ctx, user.StudyPlaceID, acceptUserID)
}

func (u *controller) Block(ctx context.Context, user global.User, blockUserID primitive.ObjectID) error {
	return u.repository.Block(ctx, user.StudyPlaceID, blockUserID)
}

func (u *controller) GetDataByCode(ctx context.Context, code string) (SignUpCode, error) {
	return u.repository.GetDataByCode(ctx, code)
}

func (u *controller) RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error {
	return u.repository.RemoveCodeByID(ctx, id)
}
