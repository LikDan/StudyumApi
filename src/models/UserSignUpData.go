package models

type UserSignUpData struct {
	Login    string `bson:"login" json:"login"`
	Name     string `bson:"name" json:"name"`
	Email    string `bson:"email" json:"email"`
	Password string `bson:"password" json:"password"`
}
