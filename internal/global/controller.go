package global

import (
	"context"
	"github.com/pkg/errors"
	"studyum/pkg/encryption"
)

var (
	NotAuthorizationError = errors.New("not authorized")
	NotValidParams        = errors.New("not valid params")
	NoPermission          = errors.New("no permission")
	ForbiddenError        = errors.New("forbidden")
	ValidationError       = errors.New("validation error")
)

type Controller interface {
	AuthViaApiToken(ctx context.Context, token string) (User, error)
}

type controller struct {
	repository Repository

	encrypt encryption.Encryption
}

func NewController(repository Repository, encrypt encryption.Encryption) Controller {
	return &controller{repository: repository, encrypt: encrypt}
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
