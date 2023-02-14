package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
)

type TrackableType string

const (
	Field TrackableType = "field"
	Array TrackableType = "array"
)

type Trackable struct {
	Collection   string
	Property     string
	DataProperty string
	Nested       string
	Value        primitive.ObjectID
	Field        reflect.Value
	Type         TrackableType
}

func DefaultTrackable(collection string) Trackable {
	return Trackable{
		Collection:   collection,
		Type:         Field,
		Property:     "_id",
		DataProperty: "appData",
	}
}
