package controllers

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
	"studyum/internal/auth/dto"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/repositories"
	"studyum/pkg/encryption"
	"studyum/pkg/hash"
	"studyum/pkg/jwt"
)

var ValidationError = errors.New("Validation error")

type Auth interface {
	Login(ctx context.Context, ip string, data dto.Login) (entities.User, jwt.TokenPair, error)

	SignUp(ctx context.Context, ip string, data dto.SignUp) (entities.User, jwt.TokenPair, error)
	SignUpStage1(ctx context.Context, user entities.User, data dto.SignUpStage1) (entities.User, error)
	SignUpStage1ViaCode(ctx context.Context, user entities.User, code string) (entities.User, error)
	SignOut(ctx context.Context, token string) error

	TerminateAll(ctx context.Context, user entities.User) error
}

type auth struct {
	sessions Sessions

	encryption encryption.Encryption

	repository     repositories.Auth
	codeRepository repositories.Code
}

func NewAuth(sessions Sessions, encryption encryption.Encryption, repository repositories.Auth, codeRepository repositories.Code) Auth {
	return &auth{sessions: sessions, encryption: encryption, repository: repository, codeRepository: codeRepository}
}

func (c *auth) Login(ctx context.Context, ip string, data dto.Login) (entities.User, jwt.TokenPair, error) {
	if len(data.Password) < 8 {
		return entities.User{}, jwt.TokenPair{}, errors.Wrap(ValidationError, "password")
	}

	user, err := c.repository.GetUserByLogin(ctx, data.Login)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	if !hash.CompareHashAndPassword(user.Password, data.Password) {
		return entities.User{}, jwt.TokenPair{}, ForbiddenErr
	}

	pair, err := c.sessions.New(ctx, user, ip)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	c.encryption.Decrypt(&user)
	return user, pair, nil
}

func (c *auth) SignUp(ctx context.Context, ip string, data dto.SignUp) (entities.User, jwt.TokenPair, error) {
	password, err := hash.Hash(data.Password)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	user := entities.User{
		Id:       primitive.NewObjectID(),
		Password: password,
		Email:    data.Email,
		Login:    data.Login,
		Name:     data.Name,
	}

	c.encryption.Encrypt(&user)
	if err = c.repository.AddUser(ctx, user); err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	pair, err := c.sessions.New(ctx, user, ip)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
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

	c.encryption.Decrypt(&user)
	return user, nil
}

func (c *auth) SignOut(ctx context.Context, token string) error {
	return c.sessions.RemoveByToken(ctx, token)
}

func (c *auth) TerminateAll(ctx context.Context, user entities.User) error {
	return c.sessions.TerminateAll(ctx, user)
}
