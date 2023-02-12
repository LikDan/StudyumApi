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
	ValidationErr = errors.New("Validation error")
)

type Controller interface {
	Create(ctx context.Context, ip string, userID string) (entities.TokenPair, error)
	CreateWithTime(ctx context.Context, ip string, userID string, d time.Duration) (entities.TokenPair, error)

	Auth(ctx context.Context, pair entities.TokenPair) (string, bool, error)

	RemoveByToken(ctx context.Context, token string) error

	LaunchCron()
	StopCron()
}

type controller struct {
	refreshExpire time.Duration
	timeout       time.Duration

	cron *cron.Cron

	jwt        base.JWT[BaseClaims]
	repository repositories.Repository
}

type BaseClaims struct{ ID string }

func NewController(cronPattern string, expire time.Duration, refreshExpire time.Duration, timeout time.Duration, secret string, repository repositories.Repository) Controller {
	jwt := base.NewJWT[BaseClaims](expire, secret)

	c := &controller{refreshExpire: refreshExpire, timeout: timeout, cron: cron.New(), jwt: jwt, repository: repository}
	_ = c.cron.AddFunc(cronPattern, c.ClearExpired)

	return c
}

func (c *controller) Create(ctx context.Context, ip string, userID string) (entities.TokenPair, error) {
	return c.CreateWithTime(ctx, ip, userID, c.jwt.GetValidTime())
}

func (c *controller) CreateWithTime(ctx context.Context, ip string, userID string, d time.Duration) (entities.TokenPair, error) {
	//839_299_365_868_340_224
	id := utils.RandomString(10)
	pair, err := c.jwt.GeneratePairWithExpireTime(BaseClaims{ID: id}, d)
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

func (c *controller) Auth(ctx context.Context, pair entities.TokenPair) (string, bool, error) {
	var id string
	needUpdate := false

	claims, ok := c.jwt.Validate(pair.Access)
	if !ok {
		var err error
		id, needUpdate, err = c.authViaRefreshToken(ctx, pair.Refresh)
		if err != nil {
			return "", false, err
		}
	} else {
		id = claims.Claims.ID
	}
	session, err := c.repository.GetByID(ctx, id)
	if err != nil {
		return "", false, err
	}

	return session.UserID, needUpdate, nil
}

func (c *controller) authViaRefreshToken(ctx context.Context, token string) (string, bool, error) {
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

func (c *controller) RemoveByToken(ctx context.Context, token string) error {
	i := strings.IndexByte(token, '|')
	if i == -1 {
		return ValidationErr
	}

	id := token[:i]
	return c.repository.RemoveByID(ctx, id)
}

func (c *controller) LaunchCron() {
	c.cron.Start()
}

func (c *controller) StopCron() {
	c.cron.Stop()
}

func (c *controller) ClearExpired() {
	logrus.Infoln("Clear expired tokens at " + time.Now().Format(time.ANSIC))

	ctx := context.Background()
	amount, err := c.repository.RemoveExpired(ctx)
	if err != nil {
		logrus.Error("Error clear expired tokens: " + err.Error())
		return
	}

	logrus.Infoln("Successfully cleared " + strconv.Itoa(amount) + " tokens")
}
