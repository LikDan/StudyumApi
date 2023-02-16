package shared

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/pkg/encryption"
)

type Shared interface {
	GetLessonByID(ctx context.Context, id primitive.ObjectID) (lesson Lesson, err error)
	GetMarkByID(ctx context.Context, id primitive.ObjectID) (mark Mark, err error)
	GetUserByID(ctx context.Context, id primitive.ObjectID) (user User, err error)
}

type shared struct {
	studyPlaceID primitive.ObjectID
	encryption   encryption.Encryption

	lessons   *mongo.Collection
	users     *mongo.Collection
	codeUsers *mongo.Collection
}

func NewShared(studyPlaceID primitive.ObjectID, encryption encryption.Encryption, lessons, users, codeUsers *mongo.Collection) Shared {
	return &shared{studyPlaceID: studyPlaceID, encryption: encryption, lessons: lessons, users: users, codeUsers: codeUsers}
}

func (r *shared) GetLessonByID(ctx context.Context, id primitive.ObjectID) (lesson Lesson, err error) {
	err = r.lessons.FindOne(ctx, bson.M{"_id": id}).Decode(&lesson)
	return
}

func (r *shared) GetMarkByID(ctx context.Context, id primitive.ObjectID) (mark Mark, err error) {
	cursor, err := r.lessons.Aggregate(ctx, bson.A{
		bson.M{"$unwind": "$marks"},
		bson.M{"$replaceRoot": bson.M{"newRoot": "$marks"}},
		bson.M{"$match": bson.M{"_id": id}},
	})
	if err != nil {
		return Mark{}, err
	}

	if !cursor.Next(ctx) {
		return Mark{}, mongo.ErrNoDocuments
	}

	err = cursor.Decode(&mark)
	return
}

func (r *shared) GetUserByID(ctx context.Context, id primitive.ObjectID) (user User, err error) {
	if err = r.users.FindOne(ctx, bson.M{"_id": id}).Decode(&user); err != nil {
		if err = r.codeUsers.FindOne(ctx, bson.M{"_id": id}).Decode(&user); err != nil {
			return User{}, err
		}
	}

	r.encryption.Decrypt(&user)
	return
}
