package handlers

import "github.com/gin-gonic/gin"

type IGeneralHandler interface {
	Uptime(ctx *gin.Context)
	GetStudyPlaces(ctx *gin.Context)
	RequestHandler(ctx *gin.Context)
}
