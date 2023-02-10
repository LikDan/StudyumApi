package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/auth/controllers"
	"studyum/internal/auth/entities"
	codes "studyum/internal/codes/controllers"
	codesEntities "studyum/internal/codes/entities"
	parser "studyum/internal/parser/handler"
	"studyum/internal/user/dto"
	entities2 "studyum/internal/user/entities"
	"studyum/internal/user/repositories"
	"studyum/pkg/encryption"
	"studyum/pkg/hash"
	jwt "studyum/pkg/jwt/controllers"
	entities3 "studyum/pkg/jwt/entities"
)

type Controller interface {
	UpdateUser(ctx context.Context, user entities.User, token, ip string, data dto.Edit) (entities.User, entities3.TokenPair, error)

	CreateCode(ctx context.Context, user entities.User, data dto.CreateCode) (entities2.SignUpCode, error)
	PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error

	GetAccept(ctx context.Context, user entities.User) ([]entities2.AcceptUser, error)
	Accept(ctx context.Context, user entities.User, acceptUserID primitive.ObjectID) error
	Block(ctx context.Context, user entities.User, blockUserID primitive.ObjectID) error

	GetDataByCode(ctx context.Context, code string) (entities2.SignUpCode, error)
	RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error
	DecryptUser(ctx context.Context, user entities.User) entities.User

	RecoverPassword(ctx context.Context, email string) error
	ResetPasswordViaCode(ctx context.Context, resetPassword dto.ResetPassword) error
}

type controller struct {
	repository      repositories.Repository
	codesController codes.Controller
	jwt             jwt.Controller

	encrypt encryption.Encryption
	parser  parser.Handler
}

func NewUserController(repository repositories.Repository, codesController codes.Controller, sessionsController jwt.Controller, encrypt encryption.Encryption, parser parser.Handler) Controller {
	return &controller{repository: repository, codesController: codesController, jwt: sessionsController, encrypt: encrypt, parser: parser}
}

func (u *controller) UpdateUser(ctx context.Context, user entities.User, token, ip string, data dto.Edit) (entities.User, entities3.TokenPair, error) {
	if data.Password != "" {
		if !user.VerifiedEmail {
			return entities.User{}, entities3.TokenPair{}, errors.Wrap(controllers.ForbiddenErr, "confirm email")
		}

		password, err := hash.Hash(data.Password)
		if err != nil {
			return entities.User{}, entities3.TokenPair{}, err
		}

		user.Password = password
	}

	u.encrypt.Decrypt(&user)

	user.Login = data.Login
	user.Email = data.Email
	user.VerifiedEmail = user.Email == data.Email
	user.PictureUrl = data.Picture

	u.encrypt.Encrypt(&user)
	if err := u.repository.UpdateUserByID(ctx, user); err != nil {
		return entities.User{}, entities3.TokenPair{}, err
	}

	if user.Email != data.Email {
		code := codesEntities.Code{
			Type:     codesEntities.Verification,
			Email:    user.Email,
			UserID:   user.Id,
			Subject:  "Confirmation code",
			To:       user.Login,
			Filename: "code.html",
		}
		if err := u.codesController.Send(ctx, code); err != nil {
			return entities.User{}, entities3.TokenPair{}, err
		}
	}

	if err := u.jwt.RemoveByToken(ctx, token); err != nil {
		return entities.User{}, entities3.TokenPair{}, err
	}

	pair, err := u.jwt.Create(ctx, ip, user.Id.Hex())
	if err != nil {
		return entities.User{}, entities3.TokenPair{}, err
	}

	u.encrypt.Decrypt(&user)
	return user, pair, nil
}

func (u *controller) PutFirebaseTokenByUserID(ctx context.Context, token primitive.ObjectID, firebaseToken string) error {
	return u.repository.PutFirebaseTokenByUserID(ctx, token, firebaseToken)
}

func (u *controller) CreateCode(ctx context.Context, user entities.User, data dto.CreateCode) (entities2.SignUpCode, error) {
	password, err := hash.Hash(data.Password)
	if err != nil {
		return entities2.SignUpCode{}, err
	}

	code := entities2.SignUpCode{
		Id:           primitive.NewObjectID(),
		Code:         data.Code,
		Name:         data.Name,
		StudyPlaceID: user.StudyPlaceID,
		Type:         data.Type,
		Typename:     data.TypeName,
		Password:     password,
	}

	u.encrypt.Encrypt(&code)
	if err := u.repository.CreateCode(ctx, code); err != nil {
		return entities2.SignUpCode{}, nil
	}

	u.encrypt.Decrypt(&code)
	return code, nil
}

func (u *controller) GetAccept(ctx context.Context, user entities.User) ([]entities2.AcceptUser, error) {
	users, err := u.repository.GetAccept(ctx, user.StudyPlaceID)
	if err != nil {
		return nil, err
	}

	u.encrypt.Decrypt(&users)
	return users, nil
}

func (u *controller) Accept(ctx context.Context, user entities.User, acceptUserID primitive.ObjectID) error {
	return u.repository.Accept(ctx, user.StudyPlaceID, acceptUserID)
}

func (u *controller) Block(ctx context.Context, user entities.User, blockUserID primitive.ObjectID) error {
	return u.repository.Block(ctx, user.StudyPlaceID, blockUserID)
}

func (u *controller) GetDataByCode(ctx context.Context, code string) (entities2.SignUpCode, error) {
	return u.repository.GetDataByCode(ctx, code)
}

func (u *controller) RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error {
	return u.repository.RemoveCodeByID(ctx, id)
}

func (u *controller) DecryptUser(_ context.Context, user entities.User) entities.User {
	u.encrypt.Decrypt(&user)
	return user
}

func (u *controller) RecoverPassword(ctx context.Context, email string) error {
	user, err := u.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	code := codesEntities.Code{
		Type:     codesEntities.PasswordReset,
		Email:    user.Email,
		UserID:   user.Id,
		Subject:  "Password recovery",
		To:       user.Login,
		Filename: "password-reset.html",
	}

	return u.codesController.Send(ctx, code)
}

func (u *controller) ResetPasswordViaCode(ctx context.Context, resetPassword dto.ResetPassword) error {
	code, err := u.codesController.Receive(ctx, codesEntities.PasswordReset, resetPassword.Code)
	if err != nil {
		return err
	}

	password, err := hash.Hash(resetPassword.NewPassword)
	if err != nil {
		return err
	}

	return u.repository.SetPasswordByUserID(ctx, code.UserID, password)
}
