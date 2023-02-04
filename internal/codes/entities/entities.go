package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Code struct {
	ID        primitive.ObjectID `bson:"_id"`
	Code      string             `bson:"code"`
	Email     string             `bson:"email"`
	UserID    primitive.ObjectID `bson:"userID"`
	Subject   string             `bson:"subject"`
	To        string             `bson:"to"`
	Filename  string             `bson:"filename"`
	Type      CodeType           `bson:"type"`
	CreatedAt time.Time          `bson:"createdAt"`
}

type CodeType string

const (
	Verification  CodeType = "VERIFICATION"
	PasswordReset CodeType = "PASSWORD_RESET"
)
