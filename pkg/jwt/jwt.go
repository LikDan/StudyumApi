package jwt

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/pkg/jwt/base"
	"studyum/pkg/jwt/controllers"
	"studyum/pkg/jwt/repositories"
	"time"
)

func NewWithMongo(cronPattern string, expire time.Duration, refreshExpire time.Duration, timeout time.Duration, secret string, sessions *mongo.Collection) controllers.Controller {
	r := repositories.NewMongo(sessions)
	return NewWithRepository(cronPattern, expire, refreshExpire, timeout, secret, r)
}

func NewWithRedis(cronPattern string, expire time.Duration, refreshExpire time.Duration, timeout time.Duration, secret string, client *redis.Client) controllers.Controller {
	r := repositories.NewRedis(client)
	return NewWithRepository(cronPattern, expire, refreshExpire, timeout, secret, r)
}

func NewWithRepository(cronPattern string, expire time.Duration, refreshExpire time.Duration, timeout time.Duration, secret string, repo repositories.Repository) controllers.Controller {
	c := controllers.NewController(cronPattern, expire, refreshExpire, timeout, secret, repo)
	return c
}

func NewBase[C any](validTime time.Duration, secret string) base.JWT[C] {
	c := base.NewJWT[C](validTime, secret)
	return c
}
