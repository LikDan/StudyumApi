package controllers

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/repositories"
	"studyum/internal/utils/jwt"
	"studyum/pkg/encryption"
	entities2 "studyum/pkg/jwt/entities"
)

type OAuth2 interface {
	GetServiceURL(ctx context.Context, service string, redirect string) (string, error)
	ReceiveUser(ctx context.Context, serviceName string, code string) (entities2.TokenPair, error)

	DecryptUser(ctx context.Context, user entities.User) entities.User
}

type oauth2 struct {
	repository repositories.OAuth2

	encryption encryption.Encryption
	jwt        jwt.JWT
}

func NewOAuth2(repository repositories.OAuth2, encryption encryption.Encryption, jwt jwt.JWT) OAuth2 {
	return &oauth2{repository: repository, encryption: encryption, jwt: jwt}
}

func (c *oauth2) GetServiceURL(ctx context.Context, serviceName string, redirect string) (string, error) {
	serviceRaw, err := c.repository.GetService(ctx, serviceName)
	if err != nil {
		return "", err
	}

	service := serviceRaw.Get()
	return service.AuthCodeURL(redirect), nil
}

func (c *oauth2) ReceiveUser(ctx context.Context, serviceName string, code string) (entities2.TokenPair, error) {
	serviceRaw, err := c.repository.GetService(ctx, serviceName)
	if err != nil {
		return entities2.TokenPair{}, err
	}
	service := serviceRaw.Get()

	token, err := service.Exchange(ctx, code)
	if err != nil {
		return entities2.TokenPair{}, err
	}

	callbackUser, err := c.repository.GetCallbackUser(ctx, service.DataUrl+token.AccessToken)
	if err != nil {
		return entities2.TokenPair{}, err
	}

	user, err := c.repository.GetUserByEmail(ctx, callbackUser.Email)
	if err != nil {
		if !errors.Is(mongo.ErrNoDocuments, err) {
			return entities2.TokenPair{}, err
		}
		user = entities.User{
			Id:            primitive.NewObjectID(),
			Email:         callbackUser.Email,
			VerifiedEmail: callbackUser.VerifiedEmail,
			Login:         callbackUser.Name,
			PictureUrl:    callbackUser.PictureUrl,
		}

		c.encryption.Encrypt(&user)
		if err = c.repository.SignUp(ctx, user); err != nil {
			return entities2.TokenPair{}, err
		}
	}

	return c.jwt.Create(ctx, "0.0.0.0", user.Id.Hex())
}

func (c *oauth2) DecryptUser(_ context.Context, user entities.User) entities.User {
	c.encryption.Decrypt(&user)
	return user
}
