package jwt

import "github.com/golang-jwt/jwt"

type TokenPair struct {
	Access  string `json:"access" bson:"access"`
	Refresh string `json:"refresh" bson:"refresh"`
}

type Claims[C any] struct {
	jwt.StandardClaims
	Claims C `json:"claims"`
}
