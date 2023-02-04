package controllers

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"studyum/internal/codes/entities"
	"studyum/internal/codes/repositories"
	"studyum/pkg/mail"
	"time"
)

var ErrForbidden = errors.New("forbidden")

type Controller interface {
	Send(ctx context.Context, code entities.Code) error
	Receive(ctx context.Context, code string) (entities.Code, error)
}

type controller struct {
	repository repositories.Repository

	mail mail.Mail

	expireTime time.Duration
	timeout    time.Duration
}

func New(repository repositories.Repository, mail mail.Mail, expireTime time.Duration, timeout time.Duration) Controller {
	rand.Seed(time.Now().Unix())
	return &controller{repository: repository, mail: mail, expireTime: expireTime, timeout: timeout}
}

// 78_364_164_096 variations
func (c *controller) generate() string {
	var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, 6)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (c *controller) sendEmail(_ context.Context, code entities.Code) error {
	data := mail.Data{"code": code.Code, "name": code.To, "expire": code.CreatedAt.Add(time.Minute * 15).Format("01-02-2006 15:04")}
	return c.mail.SendFile(code.Email, code.Subject, code.Filename, data)
}

func (c *controller) Send(ctx context.Context, code entities.Code) error {
	if tempCode, err := c.repository.GetCodeByEmail(ctx, code.Email); err == nil {
		if tempCode.CreatedAt.Add(c.timeout).After(time.Now()) {
			return ErrForbidden
		}
	}

	if tempCode, err := c.repository.GetCodeByUserID(ctx, code.UserID); err == nil {
		if tempCode.CreatedAt.Add(c.timeout).After(time.Now()) {
			return ErrForbidden
		}
	}

	if err := c.repository.DeleteAllByEmail(ctx, code.Email); err != nil {
		return err
	}

	if err := c.repository.DeleteAllByUserID(ctx, code.UserID); err != nil {
		return err
	}

	code.ID = primitive.NewObjectID()
	code.Code = c.generate()
	code.CreatedAt = time.Now()

	if err := c.repository.Create(ctx, code); err != nil {
		return err
	}

	return c.sendEmail(ctx, code)
}

func (c *controller) Receive(ctx context.Context, rawCode string) (entities.Code, error) {
	code, err := c.repository.GetCodeAndDelete(ctx, rawCode)
	if err != nil {
		return entities.Code{}, err
	}

	if code.CreatedAt.Add(c.expireTime).Before(time.Now()) {
		return entities.Code{}, ErrForbidden
	}

	return code, err
}
