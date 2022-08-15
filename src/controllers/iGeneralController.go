package controllers

import "github.com/gin-gonic/gin"

type IGeneralController interface {
	GetStudyPlaces(ctx *gin.Context)
}
