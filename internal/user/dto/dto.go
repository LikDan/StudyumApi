package dto

type Edit struct {
	Login    string `json:"login" binding:"req"`
	Email    string `json:"email" binding:"email"`
	Picture  string `json:"picture" binding:"req"`
	Password string `json:"password" binding:"min=8|eq="`
}

type CreateCode struct {
	Code     string `json:"code" binding:"req"`
	Name     string `json:"name" binding:"req"`
	Role     string `json:"role" binding:"req"`
	RoleName string `json:"roleName" binding:"req"`
	Password string `json:"password" binding:"min=8"`
}

type ResetPassword struct {
	Code        string `json:"code" binding:"req"`
	NewPassword string `json:"password" binding:"min=8"`
}
