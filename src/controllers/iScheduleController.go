package controllers

import "github.com/gin-gonic/gin"

type IScheduleController interface {
	GetSchedule(ctx *gin.Context)
	GetMySchedule(ctx *gin.Context)

	GetScheduleTypes(ctx *gin.Context)

	UpdateSchedule(ctx *gin.Context)
	UpdateGeneralSchedule(ctx *gin.Context)

	AddLesson(ctx *gin.Context)
	UpdateLesson(ctx *gin.Context)
	DeleteLesson(ctx *gin.Context)
}
