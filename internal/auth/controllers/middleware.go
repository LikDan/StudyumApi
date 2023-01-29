package controllers

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/repositories"
	"studyum/pkg/jwt"
	"time"
)

var (
	BadClaimsErr = errors.New("bad claims")
	ForbiddenErr = errors.New("forbidden")
)

type Middleware interface {
	Auth(ctx context.Context, pair jwt.TokenPair, ip string, permissions ...string) (jwt.TokenPair, bool, entities.User, error)
	MemberAuth(ctx context.Context, pair jwt.TokenPair, ip string, permissions ...string) (jwt.TokenPair, bool, entities.User, error)
	Recover(ctx context.Context, oldPair, newPair jwt.TokenPair, ip string, userID primitive.ObjectID) error

	AuthViaApiToken(ctx context.Context, token string) (entities.User, error)
}

type middleware struct {
	jwt        jwt.JWT[entities.JWTClaims]
	repository repositories.Middleware
}

func NewMiddleware(jwt jwt.JWT[entities.JWTClaims], repository repositories.Middleware) Middleware {
	return &middleware{jwt: jwt, repository: repository}
}

func (c *middleware) Recover(ctx context.Context, oldToken, newToken jwt.TokenPair, ip string, userID primitive.ObjectID) error {
	_ = c.repository.DeleteSessionByToken(ctx, newToken.Refresh)
	session := entities.Session{
		RefreshToken: oldToken.Refresh,
		IP:           ip,
		LastOnline:   time.Now(),
	}

	return c.repository.AddSession(ctx, session, userID)
}

func (c *middleware) Auth(ctx context.Context, pair jwt.TokenPair, ip string, permissions ...string) (jwt.TokenPair, bool, entities.User, error) {
	newPair := jwt.TokenPair{}

	claims, ok, update := c.jwt.Validate(pair.Access)
	if !ok {
		access, err := c.jwt.Refresh(ctx, pair.Refresh)
		if err != nil {
			return jwt.TokenPair{}, false, entities.User{}, err
		}
		claims, ok, _ = c.jwt.Validate(access)
		if !ok {
			return jwt.TokenPair{}, false, entities.User{}, BadClaimsErr
		}
		newPair.Access = access
		newPair.Refresh, err = c.jwt.GenerateRefresh()
		if err != nil {
			return jwt.TokenPair{}, false, entities.User{}, err
		}

		session := entities.Session{
			RefreshToken: newPair.Refresh,
			IP:           ip,
			LastOnline:   time.Now(),
		}

		if err = c.repository.AddSession(ctx, session, claims.Claims.ID); err != nil {
			return jwt.TokenPair{}, false, entities.User{}, err
		}
	}
	if update {
		updatedPair, err := c.jwt.GeneratePair(claims.Claims)
		if err == nil {
			newPair = updatedPair
		}
	}

	user, err := c.repository.GetUserByID(ctx, claims.Claims.ID)
	if err != nil {
		return jwt.TokenPair{}, false, entities.User{}, err
	}

	if !c.hasPermission(user, permissions) {
		return jwt.TokenPair{}, false, entities.User{}, ForbiddenErr
	}

	return newPair, newPair.Access != "", user, nil
}

func (c *middleware) MemberAuth(ctx context.Context, pair jwt.TokenPair, ip string, permissions ...string) (jwt.TokenPair, bool, entities.User, error) {
	tokenPair, shouldUpdate, user, err := c.Auth(ctx, pair, ip, permissions...)
	if err != nil {
		return jwt.TokenPair{}, false, entities.User{}, err
	}

	if !user.Accepted || user.Blocked {
		return jwt.TokenPair{}, false, entities.User{}, ForbiddenErr
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
