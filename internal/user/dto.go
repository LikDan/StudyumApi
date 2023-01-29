package user

type Edit struct {
	Login    string `json:"login" binding:"excludesall= ,required"`
	Email    string `json:"email" binding:"email"`
	Picture  string `json:"picture" binding:"excludesall= ,required"`
	Password string `json:"password" binding:"min=8|eq="`
}

type UserCreateCodeDTO struct {
	Code     string `json:"code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	TypeName string `json:"typeName" binding:"required"`
}
