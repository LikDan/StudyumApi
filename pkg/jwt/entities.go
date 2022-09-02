package jwt

import "github.com/dgrijalva/jwt-go"

type TokenPair struct {
	Access  string `json:"access" bson:"access"`
	Refresh string `json:"refresh" bson:"refresh"`
}

type Claims[C any] struct {
	jwt.StandardClaims
	Claims C `json:"claims"`
}
