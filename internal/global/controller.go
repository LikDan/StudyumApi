package global

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/pkg/encryption"
	"studyum/pkg/jwt"
	"time"
)

var (
	NotAuthorizationError = errors.New("not authorized")
	NotValidParams        = errors.New("not valid params")
	NoPermission          = errors.New("no permission")
	ForbiddenError        = errors.New("forbidden")
	ValidationError       = errors.New("validation error")
)

type Controller interface {
	Auth(ctx context.Context, token string, blockedOrNotAccepted bool, permissions ...string) (User, error)
	AuthJWTByRefreshToken(ctx context.Context, token string, ip string, blockedOrNotAccepted bool, permissions ...string) (User, jwt.TokenPair, error)
	AuthViaApiToken(ctx context.Context, token string) (User, error)

	GetClaims(ctx context.Context, refreshToken string) (error, JWTClaims)
}

type controller struct {
	repository Repository

	jwt     jwt.JWT[JWTClaims]
	encrypt encryption.Encryption
}

func NewController(jwt jwt.JWT[JWTClaims], repository Repository, encrypt encryption.Encryption) Controller {
	return &controller{repository: repository, jwt: jwt, encrypt: encrypt}
}

func (c *controller) Auth(ctx context.Context, token string, blockedOrNotAccepted bool, permissions ...string) (User, error) {
	claims, ok := c.jwt.Validate(token)
	if !ok {
		return User{}, errors.Wrap(NotAuthorizationError, "not valid token")
	}

	user_, err := c.repository.GetUserByID(ctx, claims.Claims.ID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, NotAuthorizationError
		} else {
			return User{}, err
		}
	}

	c.encrypt.Decrypt(&user_)

	isAdmin := false
	for _, permission := range user_.Permissions {
		if permission != "admin" {
			continue
		}

		isAdmin = true
	}

	if isAdmin {
		return user_, nil
	}

	for _, permission := range permissions {
		ret := true
		for _, existedPermission := range user_.Permissions {
			if permission == existedPermission {
				ret = false
				break
			}
		}

		if ret {
			return User{}, errors.Wrap(ForbiddenError, permission)
		}
	}

	if !user_.Accepted && !blockedOrNotAccepted {
		return User{}, errors.Wrap(ForbiddenError, "not accepted")
	}

	if user_.Blocked && !blockedOrNotAccepted {
		return User{}, errors.Wrap(ForbiddenError, "blocked")
	}

	return user_, nil
}

func (c *controller) UpdateJWTTokensViaNewSession(ctx context.Context, session Session) (error, jwt.TokenPair) {
	pair, err := c.jwt.RefreshPair(ctx, session.RefreshToken)
	if err != nil {
		return err, jwt.TokenPair{}
	}

	fmt.Println("Updating refresh token ", session.RefreshToken, " to ", pair.Refresh)

	old := session.RefreshToken
	session.RefreshToken = pair.Refresh
	err = c.repository.SetRefreshToken(ctx, old, session)
	return err, pair
}

func (c *controller) AuthJWTByRefreshToken(ctx context.Context, token string, ip string, blockedOrNotAccepted bool, permissions ...string) (User, jwt.TokenPair, error) {
	session := Session{
		RefreshToken: token,
		IP:           ip,
		LastOnline:   time.Now(),
	}
	err, pair := c.UpdateJWTTokensViaNewSession(ctx, session)
	if err != nil {
		return User{}, jwt.TokenPair{}, err
	}

	user_, err := c.Auth(ctx, pair.Access, blockedOrNotAccepted, permissions...)
	if err != nil {
		return User{}, jwt.TokenPair{}, err
	}

	return user_, pair, nil
}

func (c *controller) AuthViaApiToken(ctx context.Context, token string) (User, error) {
	err, studyPlace := c.repository.GetStudyPlaceByApiToken(ctx, token)
	if err != nil {
		return User{}, err
	}

	user_, err := c.repository.GetUserByID(ctx, studyPlace.AdminID)
	if err != nil {
		return User{}, err
	}

	c.encrypt.Decrypt(&user_)
	return user_, nil
}

func (c *controller) GetClaims(ctx context.Context, refreshToken string) (error, JWTClaims) {
	user_, err := c.repository.GetUserViaRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return NotAuthorizationError, JWTClaims{}
		} else {
			return err, JWTClaims{}
		}
	}

	c.encrypt.Decrypt(&user_)
	claims := JWTClaims{
		ID:            user_.Id,
		Login:         user_.Login,
		Permissions:   user_.Permissions,
		FirebaseToken: user_.FirebaseToken,
	}

	return err, claims
}
