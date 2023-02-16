package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/apps/shared"
)

type Data interface {
	Get(ctx context.Context, collection, property string, value primitive.ObjectID) (data bson.M, err error)
	GetNested(ctx context.Context, collection, arr, property string, value primitive.ObjectID) (data bson.M, err error)

	Insert(ctx context.Context, collection, property string, value primitive.ObjectID, dataProperty string, resultData shared.Data) error
	InsertNested(ctx context.Context, collection, arr, property string, value primitive.ObjectID, dataProperty string, resultData shared.Data) error
}

type data struct {
	db *mongo.Database
}

func NewData(db *mongo.Database) Data {
	return &data{db: db}
}

func (r *data) Get(ctx context.Context, collection, property string, value primitive.ObjectID) (data bson.M, err error) {
	err = r.db.Collection(collection).FindOne(ctx, bson.M{property: value}).Decode(&data)
	return
}

func (r *data) GetNested(ctx context.Context, collection, arr, property string, value primitive.ObjectID) (data bson.M, err error) {
	cursor, err := r.db.Collection(collection).Aggregate(ctx, bson.A{
		bson.M{"$unwind": "$" + arr},
		bson.M{"$replaceRoot": bson.M{"newRoot": "$" + arr}},
		bson.M{"$match": bson.M{property: value}},
	})
	if err != nil {
		return nil, err
	}

	if !cursor.Next(ctx) {
		return nil, mongo.ErrNoDocuments
	}

	err = cursor.Decode(&data)
	return
}

func (r *data) Insert(ctx context.Context, collection, property string, value primitive.ObjectID, dataProperty string, resultData shared.Data) error {
	_, err := r.db.Collection(collection).UpdateOne(ctx, bson.M{property: value}, bson.M{"$set": bson.M{dataProperty: resultData}})
	return err
}

func (r *data) InsertNested(ctx context.Context, collection, arr, property string, value primitive.ObjectID, dataProperty string, resultData shared.Data) error {
	_, err := r.db.Collection(collection).UpdateOne(ctx, bson.M{arr + "." + property: value}, bson.M{"$set": bson.M{arr + ".$." + dataProperty: resultData}})
	return err
}
