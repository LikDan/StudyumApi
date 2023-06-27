package jwt

import (
	jwt "studyum/pkg/jwt/controllers"
	"studyum/pkg/jwt/entities"
)

type JWT = jwt.Controller[Claims]

type Claims struct {
	entities.IDClaims
	UserID      string   `json:"userID" bson:"userID"`
	Login       string   `json:"login" bson:"login"`
	Name        string   `json:"name" bson:"name"`
	PictureURL  string   `json:"pictureURL" bson:"pictureURL"`
	Permissions []string `json:"permissions" bson:"permissions"`
}
