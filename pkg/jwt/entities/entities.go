package entities

import (
	"github.com/golang-jwt/jwt"
	"time"
)

type TokenPair struct {
	Access  string `json:"access" bson:"access"`
	Refresh string `json:"refresh" bson:"refresh"`
}

type IIDClaims interface {
	GetID() string
}

type IDClaims struct {
	ID string `json:"id" bson:"id"`
}

func (i IDClaims) GetID() string {
	return i.ID
}

type BaseClaims[C any] struct {
	jwt.StandardClaims
	Claims C `json:"claims"`
}

type Session struct {
	ID      string    `json:"id" bson:"_id"`
	Token   string    `json:"token" bson:"token"`
	IP      string    `json:"ip" bson:"ip"`
	UserID  string    `json:"userID" bson:"userID"`
	Expire  time.Time `json:"expire" bson:"expire"`
	Updated bool      `json:"updated" bson:"updated"`
}
