package jwt

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/golang-jwt/jwt"
	"time"
)

type JWT[C any] interface {
	Validate(token string) (Claims[C], bool)

	GeneratePair(claims C) (TokenPair, error)
	GenerateAccess(claims C) (string, error)
	GenerateRefresh() (string, error)

	RefreshPair(ctx context.Context, token string) (TokenPair, error)
	Refresh(ctx context.Context, token string) (string, error)

	SetGetClaimsFunc(fn GetClaimsByRefreshToken[C])
}

type GetClaimsByRefreshToken[C any] func(ctx context.Context, refresh string) (C, error)

type controller[C any] struct {
	validTime time.Duration
	secret    string

	getClaims GetClaimsByRefreshToken[C]
}

func New[C any](validTime time.Duration, secret string) JWT[C] {
	return &controller[C]{validTime: validTime, secret: secret}
}

func (c *controller[C]) Validate(token string) (Claims[C], bool) {
	claims := Claims[C]{}

	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.secret), nil
	})
	if err != nil {
		return Claims[C]{}, false
	}

	return claims, true
}

func (c *controller[C]) GeneratePair(claims C) (TokenPair, error) {
	access, err := c.GenerateAccess(claims)
	if err != nil {
		return TokenPair{}, err
	}

	refresh, err := c.GenerateRefresh()
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (c *controller[C]) GenerateAccess(claims C) (string, error) {
	cl := Claims[C]{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(c.validTime).Unix(),
		},
		Claims: claims,
	}
	str, err := jwt.NewWithClaims(jwt.SigningMethodHS256, cl).SignedString([]byte(c.secret))
	return str, err
}

func (c *controller[C]) GenerateRefresh() (string, error) {
	bytes := make([]byte, 128)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", bytes), nil
}

func (c *controller[C]) RefreshPair(ctx context.Context, token string) (TokenPair, error) {
	access, err := c.Refresh(ctx, token)
	if err != nil {
		return TokenPair{}, err
	}

	refresh, err := c.GenerateRefresh()
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (c *controller[C]) Refresh(ctx context.Context, token string) (string, error) {
	claims, err := c.getClaims(ctx, token)
	if err != nil {
		return "", err
	}

	return c.GenerateAccess(claims)
}

func (c *controller[C]) SetGetClaimsFunc(fn GetClaimsByRefreshToken[C]) {
	c.getClaims = fn
}
