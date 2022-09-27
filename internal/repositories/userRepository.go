package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id primitive.ObjectID) (entities.User, error)
	GetUserViaRefreshToken(ctx context.Context, refreshToken string) (entities.User, error)
	GetUserByLogin(ctx context.Context, email string) (entities.User, error)

	SignUp(ctx context.Context, user entities.User) (primitive.ObjectID, error)
	SignUpStage1(ctx context.Context, user entities.User) error

	DeleteSessionByRefreshToken(ctx context.Context, token string) error
	DeleteSessionByIP(ctx context.Context, id primitive.ObjectID, ip string) error

	UpdateUser(ctx context.Context, user entities.User) error
	UpdateUserByID(ctx context.Context, user entities.User) error

	CreateCode(ctx context.Context, code entities.SignUpCode) error

	SetRefreshToken(ctx context.Context, old string, session entities.Session) error
	AddSessionByUserID(ctx context.Context, session entities.Session, id primitive.ObjectID, sessionsAmount int) error
	RevokeToken(ctx context.Context, token string) error
	UpdateToken(ctx context.Context, id primitive.ObjectID, token string) error
	UpdateUserTokenByEmail(ctx context.Context, email, token string) error

	PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error
}

type userRepository struct {
	*Repository
}

func NewUserRepository(repository *Repository) UserRepository {
	return &userRepository{Repository: repository}
}

func (u *userRepository) GetUserByID(ctx context.Context, id primitive.ObjectID) (user entities.User, err error) {
	err = u.usersCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return
}

func (u *userRepository) GetUserViaRefreshToken(ctx context.Context, refreshToken string) (user entities.User, err error) {
	err = u.usersCollection.FindOne(ctx, bson.M{"sessions.refreshToken": refreshToken}).Decode(&user)
	return
}

func (u *userRepository) SignUp(ctx context.Context, user entities.User) (primitive.ObjectID, error) {
	if _, err := u.usersCollection.InsertOne(ctx, user); err != nil {
		return primitive.NilObjectID, err
	}

	return user.Id, nil
}

func (u *userRepository) SignUpStage1(ctx context.Context, user entities.User) error {
	_, err := u.usersCollection.UpdateByID(ctx, user.Id, bson.M{"$set": user})
	return err
}

func (u *userRepository) DeleteSessionByRefreshToken(ctx context.Context, token string) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{}, bson.M{"$pull": bson.M{"sessions": bson.M{"refreshToken": token}}})
	return err
}

func (u *userRepository) DeleteSessionByIP(ctx context.Context, id primitive.ObjectID, ip string) error {
	_, err := u.usersCollection.UpdateByID(ctx, id, bson.M{"$pull": bson.M{"sessions": bson.M{"ip": ip}}})
	return err
}

func (u *userRepository) UpdateUser(ctx context.Context, user entities.User) error {
	_, err := u.usersCollection.UpdateByID(ctx, user.Id, bson.M{"$set": user})
	return err
}

func (u *userRepository) UpdateUserByID(ctx context.Context, user entities.User) error {
	_, err := u.usersCollection.UpdateByID(ctx, user.Id, bson.M{"$set": user})
	return err
}

func (u *userRepository) SetRefreshToken(ctx context.Context, old string, session entities.Session) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"sessions.refreshToken": old}, bson.M{"$set": bson.M{"sessions.$": session}})
	return err
}

func (u *userRepository) AddSessionByUserID(ctx context.Context, session entities.Session, id primitive.ObjectID, sessionsAmount int) error {
	query := bson.M{"$addToSet": bson.M{"sessions": session}}
	if sessionsAmount == 0 {
		query = bson.M{"$set": bson.M{"sessions": bson.A{session}}}
	}

	_, err := u.usersCollection.UpdateByID(ctx, id, query)
	return err
}

func (u *userRepository) RevokeToken(ctx context.Context, token string) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"sessions.refreshToken": token}, bson.M{"$set": bson.M{"sessions": bson.A{}}})
	return err
}

func (u *userRepository) UpdateToken(ctx context.Context, id primitive.ObjectID, token string) error {
	_, err := u.usersCollection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"token": token}})
	return err
}

func (u *userRepository) UpdateUserTokenByEmail(ctx context.Context, email, token string) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": bson.M{"token": token}})
	return err
}

func (u *userRepository) GetUserByLogin(ctx context.Context, login string) (entities.User, error) {
	var user entities.User
	err := u.usersCollection.FindOne(ctx, bson.M{"login": login}).Decode(&user)

	return user, err
}

func (u *userRepository) PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error {
	_, err := u.usersCollection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"firebaseToken": firebaseToken}})
	return err
}

func (u *userRepository) CreateCode(ctx context.Context, code entities.SignUpCode) error {
	_, err := u.signUpCodesCollection.InsertOne(ctx, code)
	return err
}
