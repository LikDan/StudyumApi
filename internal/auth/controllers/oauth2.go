package controllers

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/repositories"
	"studyum/pkg/encryption"
	"studyum/pkg/jwt"
	"time"
)

type OAuth2 interface {
	GetServiceURL(ctx context.Context, service string, redirect string) (string, error)
	ReceiveUser(ctx context.Context, serviceName string, code string) (jwt.TokenPair, error)

	DecryptUser(ctx context.Context, user entities.User) entities.User
}

type oauth2 struct {
	repository         repositories.OAuth2
	sessionsRepository repositories.Sessions

	encryption encryption.Encryption
	jwt        jwt.JWT[entities.JWTClaims]
}

func NewOAuth2(repository repositories.OAuth2, sessionsRepository repositories.Sessions, encryption encryption.Encryption, jwt jwt.JWT[entities.JWTClaims]) OAuth2 {
	return &oauth2{repository: repository, sessionsRepository: sessionsRepository, encryption: encryption, jwt: jwt}
}

func (c *oauth2) GetServiceURL(ctx context.Context, serviceName string, redirect string) (string, error) {
	serviceRaw, err := c.repository.GetService(ctx, serviceName)
	if err != nil {
		return "", err
	}

	service := serviceRaw.Get()
	return service.AuthCodeURL(redirect), nil
}

func (c *oauth2) ReceiveUser(ctx context.Context, serviceName string, code string) (jwt.TokenPair, error) {
	serviceRaw, err := c.repository.GetService(ctx, serviceName)
	if err != nil {
		return jwt.TokenPair{}, err
	}
	service := serviceRaw.Get()

	token, err := service.Exchange(ctx, code)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	callbackUser, err := c.repository.GetCallbackUser(ctx, service.DataUrl+token.AccessToken)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	user, err := c.repository.GetUserByEmail(ctx, callbackUser.Email)
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
			Sessions:      make([]entities.Session, 0, 1),
		}

		c.encryption.Encrypt(&user)
		if err = c.repository.SignUp(ctx, user); err != nil {
			return jwt.TokenPair{}, err
		}
	}

	claims := entities.JWTClaims{
		ID:          user.Id,
		Login:       user.Login,
		Permissions: user.Permissions,
	}
	pair, err := c.jwt.GeneratePair(claims)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	session := entities.Session{
		RefreshToken: pair.Refresh,
		LastOnline:   time.Now(),
	}

	if err = c.sessionsRepository.Add(ctx, session, user.Id); err != nil {
		return jwt.TokenPair{}, err
	}

	return pair, nil
}

func (c *oauth2) DecryptUser(_ context.Context, user entities.User) entities.User {
	c.encryption.Decrypt(&user)
	return user
}
