package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type JWTClaims struct {
	ID            primitive.ObjectID `json:"id"`
	Login         string             `json:"login"`
	Permissions   []string           `json:"permissions"`
	FirebaseToken string             `json:"firebaseToken"`
}

type Session struct {
	RefreshToken string    `json:"-" bson:"refreshToken"`
	IP           string    `json:"ip" bson:"ip"`
	LastOnline   time.Time `json:"lastOnline" bson:"lastOnline"`
}
