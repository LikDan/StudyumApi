package models

type UserLoginData struct {
	Email    string `json:"email" bson:"email" validate:"email"`
	Password string `json:"password" bson:"password" validate:"min=8"`
}
