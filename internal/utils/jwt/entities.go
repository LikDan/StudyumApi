package jwt

import (
	jwt "studyum/pkg/jwt/controllers"
	"studyum/pkg/jwt/entities"
)

type JWT = jwt.Controller[Claims]

type Claims struct {
	entities.IDClaims
	UserID         string               `json:"userID"`
	Login          string               `json:"login"`
	PictureURL     string               `json:"pictureURL"`
	Email          string               `json:"email"`
	VerifiedEmail  bool                 `json:"verifiedEmail"`
	StudyPlaceInfo ClaimsStudyPlaceInfo `json:"studyPlaceInfo"`
}

type ClaimsStudyPlaceInfo struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	RoleName     string   `json:"roleName"`
	TuitionGroup string   `json:"tuitionGroup"`
	Permissions  []string `json:"permissions"`
	Accepted     bool     `json:"accepted"`
}
