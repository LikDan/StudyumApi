package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/entities"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id primitive.ObjectID) (entities.User, error)
	GetUserViaToken(ctx context.Context, token string, permissions ...string) (entities.User, error)
	GetUserViaRefreshToken(ctx context.Context, refreshToken string) (entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (entities.User, error)

	SignUp(ctx context.Context, user entities.User) (primitive.ObjectID, error)
	SignUpStage1(ctx context.Context, user entities.User) error

	GetUserByLogin(ctx context.Context, email string) (entities.User, error)

	UpdateUser(ctx context.Context, user entities.User) error
	UpdateUserByID(ctx context.Context, user entities.User) error

	SetRefreshToken(ctx context.Context, old string, new string) error
	SetRefreshTokenByUserID(ctx context.Context, refresh string, id primitive.ObjectID) error
	RevokeToken(ctx context.Context, token string) error
	UpdateToken(ctx context.Context, id primitive.ObjectID, token string) error
	UpdateUserTokenByEmail(ctx context.Context, email, token string) error

	PutFirebaseToken(ctx context.Context, token string, firebaseToken string) error
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

func (u *userRepository) GetUserViaToken(ctx context.Context, token string, permissions ...string) (entities.User, error) {
	var user entities.User

	filter := bson.M{"token": token}
	if len(permissions) != 0 {
		filter["permissions"] = bson.M{"$all": permissions}
	}
	err := u.usersCollection.FindOne(ctx, filter).Decode(&user)
	return user, err
}

func (u *userRepository) GetUserViaRefreshToken(ctx context.Context, refreshToken string) (user entities.User, err error) {
	err = u.usersCollection.FindOne(ctx, bson.M{"refreshToken": refreshToken}).Decode(&user)
	return
}

func (u *userRepository) SignUp(ctx context.Context, user entities.User) (primitive.ObjectID, error) {
	user.Id = primitive.NewObjectID()
	if _, err := u.usersCollection.InsertOne(ctx, user); err != nil {
		return primitive.NilObjectID, err
	}

	return user.Id, nil
}

func (u *userRepository) SignUpStage1(ctx context.Context, user entities.User) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": user.Token}, bson.M{"$set": user})
	return err
}

func (u *userRepository) GetUserByLogin(ctx context.Context, email string) (entities.User, error) {
	var user entities.User
	err := u.usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	return user, err
}

func (u *userRepository) UpdateUser(ctx context.Context, user entities.User) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": user.Token}, bson.M{"$set": user})
	return err
}

func (u *userRepository) UpdateUserByID(ctx context.Context, user entities.User) error {
	_, err := u.usersCollection.UpdateByID(ctx, user.Id, bson.M{"$set": user})
	return err
}

func (u *userRepository) SetRefreshToken(ctx context.Context, old, new string) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"refreshToken": old}, bson.M{"$set": bson.M{"refreshToken": new}})
	return err
}

func (u *userRepository) SetRefreshTokenByUserID(ctx context.Context, refresh string, id primitive.ObjectID) error {
	_, err := u.usersCollection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"refreshToken": refresh}})
	return err
}

func (u *userRepository) RevokeToken(ctx context.Context, token string) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": token}, bson.M{"$set": bson.M{"token": ""}})
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

func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	var user entities.User
	err := u.usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	return user, err
}

func (u *userRepository) PutFirebaseToken(ctx context.Context, token string, firebaseToken string) error {
	_, err := u.usersCollection.UpdateOne(ctx, bson.M{"token": token}, bson.M{"$set": bson.M{"firebaseToken": firebaseToken}})
	return err
}
