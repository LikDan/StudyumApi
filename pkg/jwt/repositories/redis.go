package repositories

import (
	"context"
	r "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"studyum/pkg/jwt/entities"
	"time"
)

type redis struct {
	client *r.Client
}

func NewRedis(client *r.Client) Repository {
	return &redis{client: client}
}

func (repo *redis) Add(ctx context.Context, session entities.Session) error {
	err := repo.client.HSet(ctx, "user:"+session.ID,
		"token", session.Token,
		"ip", session.IP,
		"userID", session.UserID,
		"expire", session.Expire.Format(time.RFC3339),
		"updated", session.Updated,
	).Err()
	if err != nil {
		return err
	}

	return repo.client.Expire(ctx, "user:"+session.ID, time.Until(session.Expire)).Err()
}

func (repo *redis) RemoveByID(ctx context.Context, id string) error {
	return repo.client.Del(ctx, "user:"+id).Err()
}

func (repo *redis) GetByID(ctx context.Context, id string) (session entities.Session, err error) {
	result, err := repo.client.HGetAll(ctx, "user:"+id).Result()
	if err != nil {
		return entities.Session{}, err
	}

	if len(result) != 5 {
		return entities.Session{}, NotValidRefreshTokenErr
	}

	session.ID = id
	session.Token = result["token"]
	session.IP = result["ip"]
	session.UserID = result["userID"]
	session.Expire, err = time.Parse(time.RFC3339, result["expire"])
	session.Updated = result["updated"] == "1"

	if err != nil {
		return entities.Session{}, err
	}

	return
}

func (repo *redis) Update(ctx context.Context, session entities.Session) error {
	return repo.Add(ctx, session)
}

func (repo *redis) RemoveExpired(context.Context) (int, error) {
	logrus.Warningln("Redis database remove expire keys automatically, so you do not need to launch cron")
	return 0, nil
}
