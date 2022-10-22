package hMongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

func Push(name string, el interface{}) bson.A {
	return bson.A{bson.M{"$set": bson.M{name: bson.M{"$ifNull": bson.A{bson.M{"$concatArrays": bson.A{"$" + name, bson.A{el}}}, bson.A{el}}}}}}
}

func AEq(el1 interface{}, el2 interface{}) bson.M {
	return bson.M{"$eq": bson.A{el1, el2}}
}

func SubstrAfter(input string, c string) string {
	return input[strings.LastIndex(input, c)+1:]
}

func Filter(input string, cond bson.M) bson.M {
	return bson.M{
		"$filter": bson.M{
			"input": "$" + input,
			"as":    SubstrAfter(input, "."),
			"cond":  cond,
		},
	}
}
