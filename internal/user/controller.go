package user

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/auth/controllers"
	"studyum/internal/auth/entities"
	parser "studyum/internal/parser/handler"
	"studyum/pkg/encryption"
	"studyum/pkg/hash"
	"studyum/pkg/jwt"
)

type Controller interface {
	UpdateUser(ctx context.Context, user entities.User, token, ip string, data Edit) (entities.User, jwt.TokenPair, error)

	SignOut(ctx context.Context, refreshToken string) error

	CreateCode(ctx context.Context, user entities.User, data UserCreateCodeDTO) (SignUpCode, error)

	RevokeToken(ctx context.Context, token string) error
	TerminateSession(ctx context.Context, user entities.User, ip string) error

	PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error

	GetAccept(ctx context.Context, user entities.User) ([]AcceptUser, error)
	Accept(ctx context.Context, user entities.User, acceptUserID primitive.ObjectID) error
	Block(ctx context.Context, user entities.User, blockUserID primitive.ObjectID) error

	GetDataByCode(ctx context.Context, code string) (SignUpCode, error)
	RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error
	DecryptUser(ctx context.Context, user entities.User) entities.User
}

type controller struct {
	repository         Repository
	sessionsController controllers.Sessions

	encrypt encryption.Encryption
	parser  parser.Handler
}

func NewUserController(repository Repository, sessionsController controllers.Sessions, encrypt encryption.Encryption, parser parser.Handler) Controller {
	return &controller{repository: repository, sessionsController: sessionsController, encrypt: encrypt, parser: parser}
}

func (u *controller) SignOut(ctx context.Context, refreshToken string) error {
	return u.repository.DeleteSessionByRefreshToken(ctx, refreshToken)
}

func (u *controller) UpdateUser(ctx context.Context, user entities.User, token, ip string, data Edit) (entities.User, jwt.TokenPair, error) {
	if data.Password != "" {
		password, err := hash.Hash(data.Password)
		if err != nil {
			return entities.User{}, jwt.TokenPair{}, err
		}

		user.Password = password
	}

	user.Login = data.Login
	user.Email = data.Email
	user.PictureUrl = data.Picture

	u.encrypt.Encrypt(&user)
	if err := u.repository.UpdateUserByID(ctx, user); err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	if err := u.sessionsController.RemoveByToken(ctx, token); err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	pair, err := u.sessionsController.New(ctx, user, ip)
	if err != nil {
		return entities.User{}, jwt.TokenPair{}, err
	}

	u.encrypt.Decrypt(&user)
	return user, pair, nil
}

func (u *controller) RevokeToken(ctx context.Context, token string) error {
	return u.repository.RevokeToken(ctx, token)
}

func (u *controller) TerminateSession(ctx context.Context, user entities.User, ip string) error {
	return u.repository.DeleteSessionByIP(ctx, user.Id, ip)
}

func (u *controller) PutFirebaseTokenByUserID(ctx context.Context, token primitive.ObjectID, firebaseToken string) error {
	return u.repository.PutFirebaseTokenByUserID(ctx, token, firebaseToken)
}

func (u *controller) CreateCode(ctx context.Context, user entities.User, data UserCreateCodeDTO) (SignUpCode, error) {
	code := SignUpCode{
		Id:           primitive.NewObjectID(),
		Code:         data.Code,
		Name:         data.Name,
		StudyPlaceID: user.StudyPlaceID,
		Type:         data.Type,
		Typename:     data.TypeName,
	}

	u.encrypt.Encrypt(&code)
	if err := u.repository.CreateCode(ctx, code); err != nil {
		return SignUpCode{}, nil
	}

	u.encrypt.Decrypt(&code)
	return code, nil
}

func (u *controller) GetAccept(ctx context.Context, user entities.User) ([]AcceptUser, error) {
	users, err := u.repository.GetAccept(ctx, user.StudyPlaceID)
	if err != nil {
		return nil, err
	}

	u.encrypt.Decrypt(&users)
	return users, nil
}

func (u *controller) Accept(ctx context.Context, user entities.User, acceptUserID primitive.ObjectID) error {
	return u.repository.Accept(ctx, user.StudyPlaceID, acceptUserID)
}

func (u *controller) Block(ctx context.Context, user entities.User, blockUserID primitive.ObjectID) error {
	return u.repository.Block(ctx, user.StudyPlaceID, blockUserID)
}

func (u *controller) GetDataByCode(ctx context.Context, code string) (SignUpCode, error) {
	return u.repository.GetDataByCode(ctx, code)
}

func (u *controller) RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error {
	return u.repository.RemoveCodeByID(ctx, id)
}

func (u *controller) DecryptUser(_ context.Context, user entities.User) entities.User {
	u.encrypt.Decrypt(&user)
	return user
}
