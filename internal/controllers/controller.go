package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/entities"
	"studyum/internal/repositories"
	"studyum/pkg/jwt"
)

var NotAuthorizationError = errors.New("not authorized")

type Controller interface {
	Auth(ctx context.Context, token string, permissions ...string) (entities.User, error)
	AuthJWT(ctx context.Context, token string, permissions ...string) (entities.User, error)

	GetClaims(ctx context.Context, refreshToken string) (error, entities.JWTClaims)
}

type controller struct {
	repository repositories.UserRepository

	jwt jwt.JWT[entities.JWTClaims]
}

func NewController(jwt jwt.JWT[entities.JWTClaims], repository repositories.UserRepository) Controller {
	return &controller{repository: repository, jwt: jwt}
}

func (c *controller) Auth(ctx context.Context, token string, permissions ...string) (entities.User, error) {
	user, err := c.repository.GetUserViaToken(ctx, token, permissions...)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return entities.User{}, NotAuthorizationError
		} else {
			return entities.User{}, err
		}
	}

	return user, err
}

func (c *controller) AuthJWT(ctx context.Context, token string, permissions ...string) (entities.User, error) {
	claims, ok := c.jwt.Validate(token)
	if !ok {
		return entities.User{}, errors.Wrap(NotAuthorizationError, "not valid token")
	}

	for _, permission := range claims.Claims.Permissions {
		ret := true
		for _, requiredPermission := range permissions {
			if permission == requiredPermission {
				ret = false
				break
			}
		}

		if ret {
			return entities.User{}, errors.Wrap(NoPermission, permission)
		}
	}

	user, err := c.repository.GetUserByID(ctx, claims.Claims.ID)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return entities.User{}, NotAuthorizationError
		} else {
			return entities.User{}, err
		}
	}

	return user, err
}

func (c *controller) GetClaims(ctx context.Context, refreshToken string) (error, entities.JWTClaims) {
	user, err := c.repository.GetUserViaRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return NotAuthorizationError, entities.JWTClaims{}
		} else {
			return err, entities.JWTClaims{}
		}
	}

	claims := entities.JWTClaims{
		Login:         user.Login,
		Permissions:   user.Permissions,
		FirebaseToken: user.FirebaseToken,
	}

	return err, claims
}
