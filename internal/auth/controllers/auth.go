package controllers

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
	"studyum/internal/auth/dto"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/repositories"
	codes "studyum/internal/codes/controllers"
	codesEntities "studyum/internal/codes/entities"
	"studyum/pkg/encryption"
	"studyum/pkg/hash"
	"studyum/pkg/jwt/controllers"
	entities2 "studyum/pkg/jwt/entities"
)

var (
	ValidationError = errors.New("Validation error")
	ErrExpired      = errors.New("Expired")
)

type Auth interface {
	UpdateByRefreshToken(ctx context.Context, token string, ip string) (entities2.TokenPair, error)

	Login(ctx context.Context, ip string, data dto.Login) (entities.User, entities2.TokenPair, error)

	SignUp(ctx context.Context, ip string, data dto.SignUp) (entities.User, entities2.TokenPair, error)
	SignUpStage1(ctx context.Context, user entities.User, data dto.SignUpStage1) (entities.User, error)
	SignUpStage1ViaCode(ctx context.Context, user entities.User, code string) (entities.User, error)
	SignOut(ctx context.Context, token string) error

	ConfirmEmail(ctx context.Context, user entities.User, code dto.VerificationCode) error
	ResendEmailCode(ctx context.Context, user entities.User) error

	TerminateAll(ctx context.Context, user entities.User) error
}

type auth struct {
	sessions controllers.Controller

	codes      codes.Controller
	encryption encryption.Encryption

	repository     repositories.Auth
	codeRepository repositories.Code
}

func NewAuth(sessions controllers.Controller, codes codes.Controller, encryption encryption.Encryption, repository repositories.Auth, codeRepository repositories.Code) Auth {
	return &auth{sessions: sessions, codes: codes, encryption: encryption, repository: repository, codeRepository: codeRepository}
}

func (c *auth) UpdateByRefreshToken(ctx context.Context, token string, ip string) (entities2.TokenPair, error) {
	return c.sessions.UpdateTokensByRefresh(ctx, token, ip)
}

func (c *auth) Login(ctx context.Context, ip string, data dto.Login) (entities.User, entities2.TokenPair, error) {
	if len(data.Password) < 8 {
		return entities.User{}, entities2.TokenPair{}, errors.Wrap(ValidationError, "password")
	}

	user, err := c.repository.GetUserByLogin(ctx, data.Login)
	if err != nil {
		return entities.User{}, entities2.TokenPair{}, err
	}

	if !hash.CompareHashAndPassword(user.Password, data.Password) {
		return entities.User{}, entities2.TokenPair{}, ForbiddenErr
	}

	pair, err := c.sessions.Create(ctx, ip, user.Id.Hex())
	if err != nil {
		return entities.User{}, entities2.TokenPair{}, err
	}

	c.encryption.Decrypt(&user)
	return user, pair, nil
}

func (c *auth) generateCode(user entities.User) codesEntities.Code {
	return codesEntities.Code{
		Type:     codesEntities.Verification,
		Email:    user.Email,
		UserID:   user.Id,
		Subject:  "Confirmation code",
		To:       user.Login,
		Filename: "code.html",
	}
}

func (c *auth) SignUp(ctx context.Context, ip string, data dto.SignUp) (entities.User, entities2.TokenPair, error) {
	var user entities.User
	var appData map[string]any
	if len(data.Password) < 8 || len(data.Email) < 5 {
		appData, _ = c.codeRepository.GetAppData(ctx, data.Code)

		code, err := c.codeRepository.GetUserByCodeAndDelete(ctx, data.Code)
		if err != nil {
			return entities.User{}, entities2.TokenPair{}, err
		}

		c.encryption.Decrypt(&code)

		user = entities.User{
			Id:           code.Id,
			Password:     code.DefaultPassword,
			Login:        data.Login,
			Type:         code.Type,
			TypeName:     code.Typename,
			StudyPlaceID: code.StudyPlaceID,
			Name:         code.Name,
			Accepted:     true,
		}
	} else {
		password, err := hash.Hash(data.Password)
		if err != nil {
			return entities.User{}, entities2.TokenPair{}, err
		}

		user = entities.User{
			Id:       primitive.NewObjectID(),
			Password: password,
			Email:    data.Email,
			Login:    data.Login,
		}
	}

	c.encryption.Encrypt(&user)
	if err := c.repository.AddUser(ctx, user); err != nil {
		return entities.User{}, entities2.TokenPair{}, err
	}

	if len(data.Password) >= 8 && len(data.Email) >= 5 {
		code := c.generateCode(user)
		if err := c.codes.Send(ctx, code); err != nil {
			return entities.User{}, entities2.TokenPair{}, err
		}
	} else {
		_ = c.repository.AddAppData(ctx, user.Id, appData)
	}

	pair, err := c.sessions.Create(ctx, ip, user.Id.Hex())
	if err != nil {
		return entities.User{}, entities2.TokenPair{}, err
	}

	c.encryption.Decrypt(&user)
	return user, pair, nil
}

func (c *auth) SignUpStage1(ctx context.Context, user entities.User, data dto.SignUpStage1) (entities.User, error) {
	user.Type = data.Type
	user.TypeName = data.TypeName
	user.StudyPlaceID = data.StudyPlaceID
	user.Blocked = false
	user.Accepted = false

	if err := c.repository.UpdateUser(ctx, user); err != nil {
		return entities.User{}, err
	}

	c.encryption.Decrypt(&user)
	return user, nil
}

func (c *auth) SignUpStage1ViaCode(ctx context.Context, user entities.User, code string) (entities.User, error) {
	appData, _ := c.codeRepository.GetAppData(ctx, code)

	data, err := c.codeRepository.GetUserByCodeAndDelete(ctx, code)
	if err != nil {
		return entities.User{}, err
	}

	user.Type = data.Type
	user.TypeName = data.Typename
	user.StudyPlaceID = data.StudyPlaceID
	user.Name = data.Name
	user.Blocked = false
	user.Accepted = true

	if err = c.repository.UpdateUser(ctx, user); err != nil {
		return entities.User{}, err
	}

	_ = c.repository.AddAppData(ctx, user.Id, appData)

	c.encryption.Decrypt(&user)
	return user, nil
}

func (c *auth) SignOut(ctx context.Context, token string) error {
	return c.sessions.RemoveByToken(ctx, token)
}

func (c *auth) ConfirmEmail(ctx context.Context, user entities.User, dto dto.VerificationCode) error {
	code, err := c.codes.Receive(ctx, codesEntities.Verification, dto.Code)
	if err != nil {
		return err
	}

	if user.VerifiedEmail || user.Email != code.Email {
		return ValidationError
	}

	return c.repository.VerifyEmail(ctx, user.Id)
}

func (c *auth) ResendEmailCode(ctx context.Context, user entities.User) error {
	code := c.generateCode(user)
	return c.codes.Send(ctx, code)
}

func (c *auth) TerminateAll(context.Context, entities.User) error {
	return nil
}
