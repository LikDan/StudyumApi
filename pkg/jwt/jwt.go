package jwt

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/pkg/jwt/base"
	"studyum/pkg/jwt/controllers"
	"studyum/pkg/jwt/entities"
	"studyum/pkg/jwt/repositories"
	"time"
)

func NewWithMongo[C entities.IIDClaims](cronPattern string, expire time.Duration, refreshExpire time.Duration, timeout time.Duration, secret string, sessions *mongo.Collection) controllers.Controller[C] {
	r := repositories.NewMongo(sessions)
	return NewWithRepository[C](cronPattern, expire, refreshExpire, timeout, secret, r)
}

func NewWithRedis[C entities.IIDClaims](cronPattern string, expire time.Duration, refreshExpire time.Duration, timeout time.Duration, secret string, client *redis.Client) controllers.Controller[C] {
	r := repositories.NewRedis(client)
	return NewWithRepository[C](cronPattern, expire, refreshExpire, timeout, secret, r)
}

func NewWithRepository[C entities.IIDClaims](cronPattern string, expire time.Duration, refreshExpire time.Duration, timeout time.Duration, secret string, repo repositories.Repository) controllers.Controller[C] {
	c := controllers.NewController[C](cronPattern, expire, refreshExpire, timeout, secret, repo)
	return c
}

func NewBase[C any](validTime time.Duration, secret string) base.JWT[C] {
	c := base.NewJWT[C](validTime, secret)
	return c
}
