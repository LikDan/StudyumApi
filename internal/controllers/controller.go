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
	AuthJWTByRefreshToken(ctx context.Context, token string, permissions ...string) (entities.User, jwt.TokenPair, error)

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

	for _, permission := range permissions {
		ret := true
		for _, existedPermission := range claims.Claims.Permissions {
			if permission == existedPermission {
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
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entities.User{}, NotAuthorizationError
		} else {
			return entities.User{}, err
		}
	}

	return user, err
}

func (c *controller) UpdateJWTTokensViaRefresh(ctx context.Context, refreshToken string) (error, jwt.TokenPair) {
	pair, err := c.jwt.RefreshPair(ctx, refreshToken)
	if err != nil {
		return err, jwt.TokenPair{}
	}

	err = c.repository.SetRefreshToken(ctx, refreshToken, pair.Refresh)
	return err, pair
}

func (c *controller) AuthJWTByRefreshToken(ctx context.Context, token string, permissions ...string) (entities.User, jwt.TokenPair, error) {
	err, pair := c.UpdateJWTTokensViaRefresh(ctx, token)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	user, err := c.AuthJWT(ctx, pair.Access, permissions...)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	return user, pair, nil
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
		ID:            user.Id,
		Login:         user.Login,
		Permissions:   user.Permissions,
		FirebaseToken: user.FirebaseToken,
	}

	return err, claims
}
