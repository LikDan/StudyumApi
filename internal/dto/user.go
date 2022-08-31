package dto

type UserLoginDTO struct {
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"min=8"`
}

type UserSignUpDTO struct {
	Login    string `json:"login" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"min=8"`
}

type UserSignUpStage1DTO struct {
	StudyPlaceId int    `json:"studyPlaceId" binding:"required"`
	Type         string `json:"type" binding:"required"`
	TypeName     string `json:"typeName" binding:"required"`
}
