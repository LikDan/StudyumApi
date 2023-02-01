package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type Login struct {
	Login    string `json:"login" binding:"excludesall= ,required"`
	Password string `json:"password" binding:"min=8"`
}

type SignUp struct {
	Login    string `json:"login" binding:"excludesall= ,required"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type SignUpWithCode struct {
	Code string `json:"code" binding:"excludesall= ,required"`
}

type VerificationCode struct {
	Code string `json:"code" binding:"excludesall= ,required"`
}

type SignUpStage1 struct {
	Name         string             `json:"name" binding:"excludesall= ,required"`
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID" binding:"excludesall= ,required"`
	Type         string             `json:"type" binding:"excludesall= ,required"`
	TypeName     string             `json:"typeName" binding:"excludesall= ,required"`
}
