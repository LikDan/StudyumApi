package controllers

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/repositories"
	"studyum/pkg/jwt/controllers"
	entities2 "studyum/pkg/jwt/entities"
)

var (
	BadClaimsErr = errors.New("bad claims")
	ForbiddenErr = errors.New("forbidden")
)

type Middleware interface {
	Auth(ctx context.Context, pair entities2.TokenPair, ip string, permissions ...string) (entities2.TokenPair, bool, entities.User, error)
	MemberAuth(ctx context.Context, pair entities2.TokenPair, ip string, permissions ...string) (entities2.TokenPair, bool, entities.User, error)

	AuthViaApiToken(ctx context.Context, token string) (entities.User, error)
}

type middleware struct {
	jwt        controllers.Controller
	repository repositories.Middleware
}

func NewMiddleware(jwt controllers.Controller, repository repositories.Middleware) Middleware {
	return &middleware{jwt: jwt, repository: repository}
}

func (c *middleware) Auth(ctx context.Context, pair entities2.TokenPair, ip string, permissions ...string) (entities2.TokenPair, bool, entities.User, error) {
	userID, update, err := c.jwt.Auth(ctx, pair)
	if err != nil {
		return entities2.TokenPair{}, false, entities.User{}, err
	}

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return entities2.TokenPair{}, false, entities.User{}, err
	}
	user, err := c.repository.GetUserByID(ctx, id)
	if err != nil {
		return entities2.TokenPair{}, false, entities.User{}, err
	}

	if update {
		pair, err = c.jwt.Create(ctx, ip, user.Id.Hex())
		if err != nil {
			return entities2.TokenPair{}, false, entities.User{}, err
		}
	}

	if !c.hasPermission(user, permissions) {
		return entities2.TokenPair{}, false, entities.User{}, ForbiddenErr
	}

	return pair, update, user, nil
}

func (c *middleware) MemberAuth(ctx context.Context, pair entities2.TokenPair, ip string, permissions ...string) (entities2.TokenPair, bool, entities.User, error) {
	tokenPair, shouldUpdate, user, err := c.Auth(ctx, pair, ip, permissions...)
	if err != nil {
		return entities2.TokenPair{}, false, entities.User{}, err
	}

	if !user.Accepted || user.Blocked {
		return entities2.TokenPair{}, false, entities.User{}, ForbiddenErr
	}

	return tokenPair, shouldUpdate, user, err
}

func (c *middleware) AuthViaApiToken(ctx context.Context, token string) (entities.User, error) {
	studyPlace, err := c.repository.GetStudyPlaceByApiToken(ctx, token)
	if err != nil {
		return entities.User{}, err
	}

	user, err := c.repository.GetUserByID(ctx, studyPlace.AdminID)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (c *middleware) hasPermission(user entities.User, permissions []string) bool {
	for _, permission := range permissions {
		found := false
		for _, uPermission := range user.Permissions {
			if uPermission == permission {
				found = true
				continue
			}
		}
		if !found {
			return false
		}
	}

	return true
}
