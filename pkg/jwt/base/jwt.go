package base

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/golang-jwt/jwt"
	jwt2 "studyum/pkg/jwt/entities"
	"time"
)

type JWT[C any] interface {
	Validate(token string) (jwt2.Claims[C], bool)

	GeneratePair(claims C) (jwt2.TokenPair, error)
	GeneratePairWithExpireTime(claims C, d time.Duration) (jwt2.TokenPair, error)

	GenerateAccess(claims C) (string, error)
	GenerateAccessWithExpireTime(claims C, d time.Duration) (string, error)

	GenerateRefresh() (string, error)

	RefreshPair(ctx context.Context, claims C) (jwt2.TokenPair, error)

	GetValidTime() time.Duration
}

type GetClaimsByRefreshToken[C any] func(ctx context.Context, refresh string) (C, error)

type j[C any] struct {
	validTime time.Duration
	secret    string
}

func NewJWT[C any](validTime time.Duration, secret string) JWT[C] {
	return &j[C]{validTime: validTime, secret: secret}
}

func (c *j[C]) Validate(token string) (jwt2.Claims[C], bool) {
	claims := jwt2.Claims[C]{}

	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.secret), nil
	})
	if err != nil {
		return jwt2.Claims[C]{}, false
	}

	return claims, true
}

func (c *j[C]) GeneratePair(claims C) (jwt2.TokenPair, error) {
	return c.GeneratePairWithExpireTime(claims, c.validTime)
}

func (c *j[C]) GeneratePairWithExpireTime(claims C, d time.Duration) (jwt2.TokenPair, error) {
	access, err := c.GenerateAccessWithExpireTime(claims, d)
	if err != nil {
		return jwt2.TokenPair{}, err
	}

	refresh, err := c.GenerateRefresh()
	if err != nil {
		return jwt2.TokenPair{}, err
	}

	return jwt2.TokenPair{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (c *j[C]) GenerateAccess(claims C) (string, error) {
	return c.GenerateAccessWithExpireTime(claims, c.validTime)
}

func (c *j[C]) GenerateAccessWithExpireTime(claims C, d time.Duration) (string, error) {
	cl := jwt2.Claims[C]{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(d).Unix(),
		},
		Claims: claims,
	}
	str, err := jwt.NewWithClaims(jwt.SigningMethodHS256, cl).SignedString([]byte(c.secret))
	return str, err
}

func (c *j[C]) GenerateRefresh() (string, error) {
	bytes := make([]byte, 128)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", bytes), nil
}

func (c *j[C]) RefreshPair(_ context.Context, claims C) (jwt2.TokenPair, error) {
	access, err := c.GenerateAccess(claims)
	if err != nil {
		return jwt2.TokenPair{}, err
	}

	refresh, err := c.GenerateRefresh()
	if err != nil {
		return jwt2.TokenPair{}, err
	}

	return jwt2.TokenPair{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (c *j[C]) GetValidTime() time.Duration {
	return c.validTime
}
