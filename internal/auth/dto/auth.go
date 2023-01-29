package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type Login struct {
	Login    string `json:"login" binding:"excludesall= ,required"`
	Password string `json:"password" binding:"min=8"`
}

type SignUp struct {
	Login    string `json:"login" binding:"excludesall= ,required"`
	Name     string `json:"name" binding:"excludesall= ,required"`
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"min=8"`
}

type SignUpWithCode struct {
	Code string `json:"code" binding:"excludesall= ,required"`
}

type SignUpStage1 struct {
	StudyPlaceID primitive.ObjectID `json:"studyPlaceID" binding:"excludesall= ,required"`
	Type         string             `json:"type" binding:"excludesall= ,required"`
	TypeName     string             `json:"typeName" binding:"excludesall= ,required"`
}
