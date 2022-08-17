package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/src/controllers"
	"studyum/src/entities"
	"studyum/src/utils"
)

type ScheduleHandler struct {
	IHandler

	controller controllers.IScheduleController

	Group *gin.RouterGroup
}

func NewScheduleHandler(authHandler IHandler, controller controllers.IScheduleController, group *gin.RouterGroup) *ScheduleHandler {
	h := &ScheduleHandler{IHandler: authHandler, controller: controller, Group: group}

	group.GET(":type/:name", h.Auth(), h.GetSchedule)
	group.GET("", h.Auth(), h.GetUserSchedule)
	group.GET("getTypes", h.Auth(), h.GetScheduleTypes)

	group.POST("", h.Auth(), h.AddLesson)
	group.PUT("", h.Auth(), h.UpdateLesson)
	group.DELETE(":id", h.Auth(), h.DeleteLesson)

	return h
}

func (s *ScheduleHandler) GetSchedule(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	type_ := ctx.Param("type")
	typeName := ctx.Param("name")

	schedule, err := s.controller.GetSchedule(ctx, type_, typeName, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

func (s *ScheduleHandler) GetUserSchedule(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	schedule, err := s.controller.GetUserSchedule(ctx, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

func (s *ScheduleHandler) GetScheduleTypes(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	types := s.controller.GetScheduleTypes(ctx, user)

	ctx.JSON(http.StatusOK, types)
}

func (s *ScheduleHandler) AddLesson(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var lesson entities.Lesson
	if err := ctx.BindJSON(&lesson); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	err := s.controller.AddLesson(ctx, lesson, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (s *ScheduleHandler) UpdateLesson(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var lesson entities.Lesson
	if err := ctx.BindJSON(&lesson); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	err := s.controller.UpdateLesson(ctx, lesson, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (s *ScheduleHandler) DeleteLesson(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)
	id := ctx.Param("id")

	err := s.controller.DeleteLesson(ctx, id, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, id)
}
