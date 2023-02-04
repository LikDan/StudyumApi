package controllers

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
	"studyum/internal/auth/dto"
	"studyum/internal/auth/entities"
	"studyum/internal/auth/repositories"
	"studyum/internal/utils"
	"studyum/pkg/encryption"
	"studyum/pkg/hash"
	"studyum/pkg/jwt"
	"studyum/pkg/mail"
	"time"
)

var (
	ValidationError = errors.New("Validation error")
	ErrExpired      = errors.New("Expired")
)

type Auth interface {
	Login(ctx context.Context, ip string, data dto.Login) (entities.User, jwt.TokenPair, error)

	SignUp(ctx context.Context, ip string, data dto.SignUp) (entities.User, jwt.TokenPair, error)
	SignUpStage1(ctx context.Context, user entities.User, data dto.SignUpStage1) (entities.User, error)
	SignUpStage1ViaCode(ctx context.Context, user entities.User, code string) (entities.User, error)
	SignOut(ctx context.Context, token string) error

	ConfirmEmail(ctx context.Context, user entities.User, code dto.VerificationCode) error
	ResendEmailCode(ctx context.Context, user entities.User) error

	TerminateAll(ctx context.Context, user entities.User) error
}

type auth struct {
	sessions Sessions
	mail     mail.Mail

	encryption encryption.Encryption

	repository                  repositories.Auth
	codeRepository              repositories.Code
	verificationsCodeRepository repositories.VerificationCodes
}

func NewAuth(sessions Sessions, mail mail.Mail, encryption encryption.Encryption, repository repositories.Auth, codeRepository repositories.Code, verificationsCodeRepository repositories.VerificationCodes) Auth {
	return &auth{sessions: sessions, mail: mail, encryption: encryption, repository: repository, codeRepository: codeRepository, verificationsCodeRepository: verificationsCodeRepository}
}

func (c *auth) Login(ctx context.Context, ip string, data dto.Login) (entities.User, jwt.TokenPair, error) {
	if len(data.Password) < 8 {
		return entities.User{}, jwt.TokenPair{}, errors.Wrap(ValidationError, "password")
	}

	user, err := c.repository.GetUserByLogin(ctx, data.Login)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	if !hash.CompareHashAndPassword(user.Password, data.Password) {
		return entities.User{}, jwt.TokenPair{}, ForbiddenErr
	}

	pair, err := c.sessions.New(ctx, user, ip)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	c.encryption.Decrypt(&user)
	return user, pair, nil
}

func (c *auth) sendCodeEmail(_ context.Context, name string, code entities.VerificationCode) error {
	emailData := mail.Data{"code": code.Code, "name": name, "expire": code.CreatedAt.Add(time.Minute * 15).Format("01-02-2006 15:04")}
	return c.mail.SendFile(code.Email, "Authorization code", "code.html", emailData)
}

func (c *auth) generateCode(ctx context.Context, userID primitive.ObjectID, email string) (entities.VerificationCode, error) {
	code := entities.VerificationCode{
		Code:      utils.RandomCode(6),
		Email:     email,
		CreatedAt: time.Now(),
		UserID:    userID,
	}

	if err := c.verificationsCodeRepository.Create(ctx, code); err != nil {
		return entities.VerificationCode{}, err
	}

	return code, nil
}

func (c *auth) SignUp(ctx context.Context, ip string, data dto.SignUp) (entities.User, jwt.TokenPair, error) {
	var user entities.User
	if len(data.Password) < 8 || len(data.Email) < 5 {
		code, err := c.codeRepository.GetUserByCodeAndDelete(ctx, data.Code)
		if err != nil {
			return entities.User{}, jwt.TokenPair{}, err
		}

		c.encryption.Decrypt(&code)

		user = entities.User{
			Id:           code.Id,
			Password:     code.DefaultPassword,
			Login:        data.Login,
			Type:         code.Type,
			TypeName:     code.Typename,
			StudyPlaceID: code.StudyPlaceID,
			Name:         code.Name,
			Accepted:     true,
		}
	} else {
		password, err := hash.Hash(data.Password)
		if err != nil {
			return entities.User{}, jwt.TokenPair{}, err
		}

		user = entities.User{
			Id:       primitive.NewObjectID(),
			Password: password,
			Email:    data.Email,
			Login:    data.Login,
		}

		code, err := c.generateCode(ctx, user.Id, data.Email)
		if err != nil {
			return entities.User{}, jwt.TokenPair{}, err
		}

		if err = c.sendCodeEmail(ctx, user.Login, code); err != nil {
			return entities.User{}, jwt.TokenPair{}, err
		}
	}

	c.encryption.Encrypt(&user)
	if err := c.repository.AddUser(ctx, user); err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	pair, err := c.sessions.New(ctx, user, ip)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	c.encryption.Decrypt(&user)
	return user, pair, nil
}

func (c *auth) SignUpStage1(ctx context.Context, user entities.User, data dto.SignUpStage1) (entities.User, error) {
	user.Type = data.Type
	user.TypeName = data.TypeName
	user.StudyPlaceID = data.StudyPlaceID
	user.Blocked = false
	user.Accepted = false

	if err := c.repository.UpdateUser(ctx, user); err != nil {
		return entities.User{}, err
	}

	c.encryption.Decrypt(&user)
	return user, nil
}

func (c *auth) SignUpStage1ViaCode(ctx context.Context, user entities.User, code string) (entities.User, error) {
	data, err := c.codeRepository.GetUserByCodeAndDelete(ctx, code)
	if err != nil {
		return entities.User{}, err
	}

	user.Type = data.Type
	user.TypeName = data.Typename
	user.StudyPlaceID = data.StudyPlaceID
	user.Name = data.Name
	user.Blocked = false
	user.Accepted = true

	if err = c.repository.UpdateUser(ctx, user); err != nil {
		return entities.User{}, err
	}

	c.encryption.Decrypt(&user)
	return user, nil
}

func (c *auth) SignOut(ctx context.Context, token string) error {
	return c.sessions.RemoveByToken(ctx, token)
}

func (c *auth) ConfirmEmail(ctx context.Context, user entities.User, code dto.VerificationCode) error {
	verificationCode, err := c.verificationsCodeRepository.GetCodeAndDelete(ctx, code.Code)
	if err != nil {
		return err
	}

	if verificationCode.CreatedAt.Add(time.Minute * 15).Before(time.Now()) {
		return errors.Wrap(ErrExpired, "code")
	}

	if user.VerifiedEmail || user.Email != verificationCode.Email || user.Id != verificationCode.UserID {
		return ValidationError
	}

	return c.repository.VerifyEmail(ctx, user.Id)
}

func (c *auth) ResendEmailCode(ctx context.Context, user entities.User) error {
	code, err := c.verificationsCodeRepository.GetCodeByEmail(ctx, user.Email)
	if err != nil {
		return err
	}

	if code.CreatedAt.Add(time.Minute).After(time.Now()) && code.UserID == user.Id {
		return errors.Wrap(ForbiddenErr, "too many requests")
	}

	if err = c.verificationsCodeRepository.DeleteAllByEmail(ctx, user.Email); err != nil {
		return err
	}

	code, err = c.generateCode(ctx, user.Id, user.Email)
	if err != nil {
		return err
	}

	return c.sendCodeEmail(ctx, user.Login, code)
}

func (c *auth) TerminateAll(ctx context.Context, user entities.User) error {
	return c.sessions.TerminateAll(ctx, user)
}
