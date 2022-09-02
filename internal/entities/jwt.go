package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type JWTClaims struct {
	ID            primitive.ObjectID `json:"id"`
	Login         string             `json:"login"`
	Permissions   []string           `json:"permissions"`
	FirebaseToken string             `json:"firebaseToken"`
}
