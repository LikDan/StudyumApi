package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type Login struct {
	Login    string `json:"login" binding:"req"`
	Password string `json:"password" binding:"min=8"`
}

type SignUp struct {
	Login    string `json:"login" binding:"req"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type SignUpWithCode struct {
	Code string `json:"code" binding:"req"`
}

type VerificationCode struct {
	Code string `json:"code" binding:"req"`
}

type SignUpStage1 struct {
	Name         string             `json:"name" binding:"req"`
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID"`
	Role         string             `json:"role" binding:"req"`
	RoleName     string             `json:"roleName" binding:"req"`
}
