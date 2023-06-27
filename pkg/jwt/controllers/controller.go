package controllers

import (
	"context"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"studyum/pkg/jwt/base"
	"studyum/pkg/jwt/entities"
	"studyum/pkg/jwt/repositories"
	"studyum/pkg/jwt/utils"
	"time"
)

var (
	ValidationErr   = errors.New("Validation error")
	RefreshTokenErr = errors.New("access token has expired")
)

type Controller[C entities.IIDClaims] interface {
	Create(ctx context.Context, ip string, userID string) (entities.TokenPair, error)
	CreateWithTime(ctx context.Context, ip string, userID string, d time.Duration) (entities.TokenPair, error)

	Auth(ctx context.Context, pair entities.TokenPair) (string, bool, error)

	RemoveByToken(ctx context.Context, token string) error
	UpdateTokensByRefresh(ctx context.Context, token string, ip string) (entities.TokenPair, error)

	SetCreateClaimsFunc(func(ctx context.Context, id, userID string) (C, error))

	LaunchCron()
	StopCron()
}

type controller[C entities.IIDClaims] struct {
	refreshExpire time.Duration
	timeout       time.Duration

	cron *cron.Cron

	jwt        base.JWT[C]
	repository repositories.Repository

	createClaimsFunc func(ctx context.Context, id, userID string) (C, error)
}

func NewController[C entities.IIDClaims](cronPattern string, expire, refreshExpire, timeout time.Duration, secret string, repository repositories.Repository) Controller[C] {
	return NewControllerWithCreateClaimsFunc[C](cronPattern, expire, refreshExpire, timeout, secret, repository, nil)
}

func NewControllerWithCreateClaimsFunc[C entities.IIDClaims](cronPattern string, expire, refreshExpire, timeout time.Duration, secret string, repository repositories.Repository, createClaimsFunc func(ctx context.Context, id, userID string) (C, error)) Controller[C] {
	jwt := base.NewJWT[C](expire, secret)

	c := &controller[C]{refreshExpire: refreshExpire, timeout: timeout, cron: cron.New(), jwt: jwt, repository: repository, createClaimsFunc: createClaimsFunc}
	_ = c.cron.AddFunc(cronPattern, c.ClearExpired)

	return c
}

func NewControllerWithSimpleClaims(cronPattern string, expire, refreshExpire, timeout time.Duration, secret string, repository repositories.Repository) Controller[entities.IIDClaims] {
	return NewControllerWithCreateClaimsFunc[entities.IIDClaims](cronPattern, expire, refreshExpire, timeout, secret, repository, func(ctx context.Context, id, userID string) (entities.IIDClaims, error) {
		return entities.IDClaims{ID: id}, nil
	})
}

func (c *controller[C]) Create(ctx context.Context, ip string, userID string) (entities.TokenPair, error) {
	return c.CreateWithTime(ctx, ip, userID, c.jwt.GetValidTime())
}

func (c *controller[C]) CreateWithTime(ctx context.Context, ip string, userID string, d time.Duration) (entities.TokenPair, error) {
	//839_299_365_868_340_224
	id := utils.RandomString(10)

	if c.createClaimsFunc == nil {
		return entities.TokenPair{}, errors.New("createClaimsFunc is nil")
	}
	claims, err := c.createClaimsFunc(ctx, id, userID)

	if err != nil {
		return entities.TokenPair{}, err
	}

	pair, err := c.jwt.GeneratePairWithExpireTime(claims, d)
	if err != nil {
		return entities.TokenPair{}, err
	}

	pair.Refresh = id + "|" + pair.Refresh

	session := entities.Session{
		ID:      id,
		Token:   pair.Refresh,
		IP:      ip,
		UserID:  userID,
		Expire:  time.Now().Add(c.refreshExpire),
		Updated: false,
	}
	if err = c.repository.Add(ctx, session); err != nil {
		return entities.TokenPair{}, err
	}

	return pair, nil
}

func (c *controller[C]) Auth(ctx context.Context, pair entities.TokenPair) (string, bool, error) {
	var id string
	needUpdate := false

	claims, ok := c.jwt.Validate(pair.Access)
	if !ok {
		var err error
		id, needUpdate, err = c.authViaRefreshToken(ctx, pair.Refresh)
		if err != nil {
			return "", false, RefreshTokenErr
		}
	} else {
		id = claims.Claims.GetID()
	}
	session, err := c.repository.GetByID(ctx, id)
	if err != nil {
		return "", false, err
	}

	return session.UserID, needUpdate, nil
}

func (c *controller[C]) authViaRefreshToken(ctx context.Context, token string) (string, bool, error) {
	i := strings.IndexByte(token, '|')
	if i == -1 {
		return "", false, ValidationErr
	}

	id := token[:i]
	session, err := c.repository.GetByID(ctx, id)
	if err != nil {
		return "", false, err
	}

	if session.Token != token {
		return "", false, ValidationErr
	}

	if time.Now().After(session.Expire) {
		return "", false, ValidationErr
	}

	if session.Updated {
		return session.ID, false, nil
	}

	session.Updated = true
	session.Expire = time.Now().Add(c.timeout)
	if err = c.repository.Update(ctx, session); err != nil {
		return "", false, err
	}

	return session.ID, true, nil
}

func (c *controller[C]) RemoveByToken(ctx context.Context, token string) error {
	i := strings.IndexByte(token, '|')
	if i == -1 {
		return ValidationErr
	}

	id := token[:i]
	return c.repository.RemoveByID(ctx, id)
}

func (c *controller[C]) UpdateTokensByRefresh(ctx context.Context, refresh string, ip string) (entities.TokenPair, error) {
	id, _, err := c.authViaRefreshToken(ctx, refresh)
	if err != nil {
		return entities.TokenPair{}, err
	}
	session, err := c.repository.GetByID(ctx, id)
	if err != nil {
		return entities.TokenPair{}, err
	}

	return c.Create(ctx, ip, session.UserID)
}

func (c *controller[C]) SetCreateClaimsFunc(f func(ctx context.Context, id, userID string) (C, error)) {
	c.createClaimsFunc = f
}

func (c *controller[C]) LaunchCron() {
	c.cron.Start()
}

func (c *controller[C]) StopCron() {
	c.cron.Stop()
}

func (c *controller[C]) ClearExpired() {
	logrus.Infoln("Clear expired tokens at " + time.Now().Format(time.ANSIC))

	ctx := context.Background()
	amount, err := c.repository.RemoveExpired(ctx)
	if err != nil {
		logrus.Error("Error clear expired tokens: " + err.Error())
		return
	}

	logrus.Infoln("Successfully cleared " + strconv.Itoa(amount) + " tokens")
}
