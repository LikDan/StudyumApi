package user

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/internal/auth/entities"
	"studyum/internal/global"
)

type Repository interface {
	GetUserByID(ctx context.Context, id primitive.ObjectID) (entities.User, error)
	GetUserViaRefreshToken(ctx context.Context, refreshToken string) (entities.User, error)
	GetUserByLogin(ctx context.Context, email string) (entities.User, error)

	SignUp(ctx context.Context, user entities.User) (primitive.ObjectID, error)
	SignUpStage1(ctx context.Context, user entities.User) error

	DeleteSessionByRefreshToken(ctx context.Context, token string) error
	DeleteSessionByIP(ctx context.Context, id primitive.ObjectID, ip string) error

	UpdateUser(ctx context.Context, user entities.User) error
	UpdateUserByID(ctx context.Context, user entities.User) error

	CreateCode(ctx context.Context, code SignUpCode) error

	SetRefreshToken(ctx context.Context, old string, session entities.Session) error
	AddSessionByUserID(ctx context.Context, session entities.Session, id primitive.ObjectID, sessionsAmount int) error
	RevokeToken(ctx context.Context, token string) error
	UpdateToken(ctx context.Context, id primitive.ObjectID, token string) error
	UpdateUserTokenByEmail(ctx context.Context, email, token string) error

	PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error

	GetAccept(ctx context.Context, studyPlaceID primitive.ObjectID) ([]AcceptUser, error)
	Accept(ctx context.Context, studyPlaceID primitive.ObjectID, userID primitive.ObjectID) error
	Block(ctx context.Context, studyPlaceID primitive.ObjectID, userID primitive.ObjectID) error

	GetDataByCode(ctx context.Context, code string) (SignUpCode, error)
	RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error
}

type repository struct {
	*global.Repository
}

func NewUserRepository(r *global.Repository) Repository {
	return &repository{Repository: r}
}

func (u *repository) GetUserByID(ctx context.Context, id primitive.ObjectID) (user entities.User, err error) {
	err = u.UsersCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return
}

func (u *repository) GetUserViaRefreshToken(ctx context.Context, refreshToken string) (user entities.User, err error) {
	err = u.UsersCollection.FindOne(ctx, bson.M{"sessions.refreshToken": refreshToken}).Decode(&user)
	return
}

func (u *repository) SignUp(ctx context.Context, user entities.User) (primitive.ObjectID, error) {
	if _, err := u.UsersCollection.InsertOne(ctx, user); err != nil {
		return primitive.NilObjectID, err
	}

	return user.Id, nil
}

func (u *repository) SignUpStage1(ctx context.Context, user entities.User) error {
	_, err := u.UsersCollection.UpdateByID(ctx, user.Id, bson.M{"$set": user})
	return err
}

func (u *repository) DeleteSessionByRefreshToken(ctx context.Context, token string) error {
	_, err := u.UsersCollection.UpdateOne(ctx, bson.M{}, bson.M{"$pull": bson.M{"sessions": bson.M{"refreshToken": token}}})
	return err
}

func (u *repository) DeleteSessionByIP(ctx context.Context, id primitive.ObjectID, ip string) error {
	_, err := u.UsersCollection.UpdateByID(ctx, id, bson.M{"$pull": bson.M{"sessions": bson.M{"ip": ip}}})
	return err
}

func (u *repository) UpdateUser(ctx context.Context, user entities.User) error {
	_, err := u.UsersCollection.UpdateByID(ctx, user.Id, bson.M{"$set": user})
	return err
}

func (u *repository) UpdateUserByID(ctx context.Context, user entities.User) error {
	_, err := u.UsersCollection.UpdateByID(ctx, user.Id, bson.M{"$set": user})
	return err
}

func (u *repository) SetRefreshToken(ctx context.Context, old string, session entities.Session) error {
	_, err := u.UsersCollection.UpdateOne(ctx, bson.M{"sessions.refreshToken": old}, bson.M{"$set": bson.M{"sessions.$": session}})
	return err
}

func (u *repository) AddSessionByUserID(ctx context.Context, session entities.Session, id primitive.ObjectID, sessionsAmount int) error {
	query := bson.M{"$addToSet": bson.M{"sessions": session}}
	if sessionsAmount == 0 {
		query = bson.M{"$set": bson.M{"sessions": bson.A{session}}}
	}

	_, err := u.UsersCollection.UpdateByID(ctx, id, query)
	return err
}

func (u *repository) RevokeToken(ctx context.Context, token string) error {
	_, err := u.UsersCollection.UpdateOne(ctx, bson.M{"sessions.refreshToken": token}, bson.M{"$set": bson.M{"sessions": bson.A{}}})
	return err
}

func (u *repository) UpdateToken(ctx context.Context, id primitive.ObjectID, token string) error {
	_, err := u.UsersCollection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"token": token}})
	return err
}

func (u *repository) UpdateUserTokenByEmail(ctx context.Context, email, token string) error {
	_, err := u.UsersCollection.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": bson.M{"token": token}})
	return err
}

func (u *repository) GetUserByLogin(ctx context.Context, login string) (entities.User, error) {
	var user entities.User
	err := u.UsersCollection.FindOne(ctx, bson.M{"login": login}).Decode(&user)

	return user, err
}

func (u *repository) PutFirebaseTokenByUserID(ctx context.Context, id primitive.ObjectID, firebaseToken string) error {
	_, err := u.UsersCollection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"firebaseToken": firebaseToken}})
	return err
}

func (u *repository) CreateCode(ctx context.Context, code SignUpCode) error {
	_, err := u.SignUpCodesCollection.InsertOne(ctx, code)
	return err
}

func (u *repository) GetAccept(ctx context.Context, studyPlaceID primitive.ObjectID) (users []AcceptUser, err error) {
	cursor, err := u.UsersCollection.Find(ctx, bson.M{"studyPlaceID": studyPlaceID, "accepted": false})
	if err != nil {
		return
	}

	err = cursor.All(ctx, &users)
	return
}

func (u *repository) Accept(ctx context.Context, studyPlaceID primitive.ObjectID, userID primitive.ObjectID) error {
	_, err := u.UsersCollection.UpdateOne(ctx, bson.M{"studyPlaceID": studyPlaceID, "_id": userID}, bson.M{"$set": bson.M{"accepted": true}})
	return err
}

func (u *repository) Block(ctx context.Context, studyPlaceID primitive.ObjectID, userID primitive.ObjectID) error {
	_, err := u.UsersCollection.UpdateOne(ctx, bson.M{"studyPlaceID": studyPlaceID, "_id": userID}, bson.M{"$set": bson.M{"blocked": true}})
	return err
}

func (u *repository) GetDataByCode(ctx context.Context, code string) (data SignUpCode, err error) {
	err = u.SignUpCodesCollection.FindOne(ctx, bson.M{"code": code}).Decode(&data)
	return
}

func (u *repository) RemoveCodeByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := u.SignUpCodesCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
