package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserLoginDTO struct {
	Login    string `json:"login" binding:"excludesall= ,required"`
	Password string `json:"password" binding:"min=8"`
}

type UserSignUpDTO struct {
	Login    string `json:"login" binding:"excludesall= ,required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"min=8"`
}

type UserSignUpWithCodeDTO struct {
	Code     string `json:"code" binding:"required"`
	Login    string `json:"login" binding:"required"`
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"min=8"`
}

type EditUserDTO struct {
	Login    string `json:"login" binding:"excludesall= ,required"`
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"min=8|eq="`
}

type UserSignUpStage1DTO struct {
	StudyPlaceId primitive.ObjectID `json:"studyPlaceId" binding:"required"`
	Type         string             `json:"type" binding:"required"`
	TypeName     string             `json:"typeName" binding:"required"`
}

type UserCreateCodeDTO struct {
	Code     string `json:"code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	TypeName string `json:"typeName" binding:"required"`
}
