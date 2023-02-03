package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type VerificationCode struct {
	Code      string             `json:"code" bson:"code"`
	Email     string             `json:"email" bson:"email"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UserID    primitive.ObjectID `json:"userID" bson:"userID"`
}
