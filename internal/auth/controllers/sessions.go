package controllers

import (
	"context"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/repositories"
	"studyum/pkg/jwt"
	"time"
)

type Sessions interface {
	New(ctx context.Context, user entities.User, ip string) (jwt.TokenPair, error)
	RemoveByToken(ctx context.Context, token string) error
}

type sessions struct {
	jwt        jwt.JWT[entities.JWTClaims]
	repository repositories.Sessions
}

func NewSessions(jwt jwt.JWT[entities.JWTClaims], repository repositories.Sessions) Sessions {
	s := &sessions{jwt: jwt, repository: repository}
	s.jwt.SetGetClaimsFunc(s.getClaimsByToken)
	return s
}

func (c *sessions) getClaimsByToken(ctx context.Context, token string) (entities.JWTClaims, error) {
	user, err := c.repository.GetUserByToken(ctx, token)
	if err != nil {
		return entities.JWTClaims{}, err
	}

	if err = c.repository.RemoveByToken(ctx, token); err != nil {
		return entities.JWTClaims{}, err
	}

	return entities.JWTClaims{
		ID:          user.Id,
		Login:       user.Login,
		Permissions: user.Permissions,
	}, nil
}

func (c *sessions) New(ctx context.Context, user entities.User, ip string) (jwt.TokenPair, error) {
	claims := entities.JWTClaims{
		ID:          user.Id,
		Login:       user.Login,
		Permissions: user.Permissions,
	}

	tokenPair, err := c.jwt.GeneratePair(claims)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	session := entities.Session{
		RefreshToken: tokenPair.Refresh,
		IP:           ip,
		LastOnline:   time.Now(),
	}

	if err = c.repository.Add(ctx, session, user.Id); err != nil {
		return jwt.TokenPair{}, err
	}

	return tokenPair, err
}

func (c *sessions) RemoveByToken(ctx context.Context, token string) error {
	return c.repository.RemoveByToken(ctx, token)
}
