package dto

type UserLoginDTO struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"min=8"`
}

type UserSignUpDTO struct {
	Login    string `bson:"login" json:"login"`
	Name     string `bson:"name" json:"name"`
	Email    string `bson:"email" json:"email" validate:"email"`
	Password string `bson:"password" json:"password"`
}

type UserSignUpStage1DTO struct {
	StudyPlaceId int    `bson:"studyPlaceId" json:"studyPlaceId"`
	Type         string `bson:"type" json:"type"`
	TypeName     string `bson:"typeName" json:"typeName"`
}
