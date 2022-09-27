package controllers

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/entities"
	"studyum/internal/repositories"
	"studyum/pkg/encryption"
	"studyum/pkg/jwt"
	"time"
)

var NotAuthorizationError = errors.New("not authorized")
var ForbiddenError = errors.New("forbidden")

type Controller interface {
	Auth(ctx context.Context, token string, permissions ...string) (entities.User, error)
	AuthJWTByRefreshToken(ctx context.Context, token string, ip string, permissions ...string) (entities.User, jwt.TokenPair, error)
	AuthViaApiToken(ctx context.Context, token string) (entities.User, error)

	GetClaims(ctx context.Context, refreshToken string) (error, entities.JWTClaims)
}

type controller struct {
	userRepository    repositories.UserRepository
	generalRepository repositories.GeneralRepository

	jwt     jwt.JWT[entities.JWTClaims]
	encrypt encryption.Encryption
}

func NewController(jwt jwt.JWT[entities.JWTClaims], userRepository repositories.UserRepository, generalRepository repositories.GeneralRepository, encrypt encryption.Encryption) Controller {
	return &controller{userRepository: userRepository, generalRepository: generalRepository, jwt: jwt, encrypt: encrypt}
}

func (c *controller) Auth(ctx context.Context, token string, permissions ...string) (entities.User, error) {
	claims, ok := c.jwt.Validate(token)
	if !ok {
		return entities.User{}, errors.Wrap(NotAuthorizationError, "not valid token")
	}

	user, err := c.userRepository.GetUserByID(ctx, claims.Claims.ID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entities.User{}, NotAuthorizationError
		} else {
			return entities.User{}, err
		}
	}

	c.encrypt.Decrypt(&user)

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

	if !user.Accepted {
		return entities.User{}, errors.Wrap(ForbiddenError, "not accepted")
	}

	if user.Blocked {
		return entities.User{}, errors.Wrap(ForbiddenError, "blocked")
	}

	return user, nil
}

func (c *controller) UpdateJWTTokensViaNewSession(ctx context.Context, session entities.Session) (error, jwt.TokenPair) {
	pair, err := c.jwt.RefreshPair(ctx, session.RefreshToken)
	if err != nil {
		return err, jwt.TokenPair{}
	}

	old := session.RefreshToken
	session.RefreshToken = pair.Refresh
	err = c.userRepository.SetRefreshToken(ctx, old, session)
	return err, pair
}

func (c *controller) AuthJWTByRefreshToken(ctx context.Context, token string, ip string, permissions ...string) (entities.User, jwt.TokenPair, error) {
	session := entities.Session{
		RefreshToken: token,
		IP:           ip,
		LastOnline:   time.Now(),
	}
	err, pair := c.UpdateJWTTokensViaNewSession(ctx, session)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	user, err := c.Auth(ctx, pair.Access, permissions...)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	return user, pair, nil
}

func (c *controller) AuthViaApiToken(ctx context.Context, token string) (entities.User, error) {
	err, studyPlace := c.generalRepository.GetStudyPlaceByApiToken(ctx, token)
	if err != nil {
		return entities.User{}, err
	}

	user, err := c.userRepository.GetUserByID(ctx, studyPlace.AdminID)
	if err != nil {
		return entities.User{}, err
	}

	c.encrypt.Decrypt(&user)
	return user, nil
}

func (c *controller) GetClaims(ctx context.Context, refreshToken string) (error, entities.JWTClaims) {
	user, err := c.userRepository.GetUserViaRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return NotAuthorizationError, entities.JWTClaims{}
		} else {
			return err, entities.JWTClaims{}
		}
	}

	c.encrypt.Decrypt(&user)
	claims := entities.JWTClaims{
		ID:            user.Id,
		Login:         user.Login,
		Permissions:   user.Permissions,
		FirebaseToken: user.FirebaseToken,
	}

	return err, claims
}
