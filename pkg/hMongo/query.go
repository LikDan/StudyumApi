package hMongo

import "go.mongodb.org/mongo-driver/bson"

func Push(name string, el interface{}) bson.A {
	return bson.A{bson.M{"$set": bson.M{name: bson.M{"$ifNull": bson.A{bson.M{"$concatArrays": bson.A{"$" + name, bson.A{el}}}, bson.A{el}}}}}}
}
