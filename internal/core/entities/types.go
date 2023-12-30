package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Subject struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	Subject string             `json:"subject" bson:"subject"`
}

type Group struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Group string             `json:"group" bson:"group"`
}

type Teacher struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	Teacher string             `json:"teacher" bson:"teacher"`
}

type Room struct {
	ID   primitive.ObjectID `json:"id" bson:"_id"`
	Room string             `json:"room" bson:"room"`
}
