package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/internal/controllers"
	"studyum/internal/entities"
	"studyum/internal/utils"
)

type JournalHandler struct {
	IHandler

	controller controllers.IJournalController

	Group *gin.RouterGroup
}

func NewJournalHandler(authHandler IHandler, controller controllers.IJournalController, group *gin.RouterGroup) *JournalHandler {
	h := &JournalHandler{IHandler: authHandler, controller: controller, Group: group}

	group.GET("/options", h.Auth(), h.GetJournalAvailableOptions)
	group.GET("/:group/:subject/:teacher", h.Auth(), h.GetJournal)
	group.GET("", h.Auth(), h.GetUserJournal)

	mark := group.Group("/mark", h.Auth())
	{
		mark.POST("", h.AddMark)
		mark.GET("", h.GetMark)
		mark.PUT("", h.UpdateMark)
		mark.DELETE("", h.DeleteMark)
	}
	return h
}

func (j *JournalHandler) GetJournalAvailableOptions(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	options, err := j.controller.GetJournalAvailableOptions(ctx, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, options)
}

func (j *JournalHandler) GetJournal(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	group := ctx.Param("group")
	subject := ctx.Param("subject")
	teacher := ctx.Param("teacher")

	journal, err := j.controller.GetJournal(ctx, group, subject, teacher, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (j *JournalHandler) GetUserJournal(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	journal, err := j.controller.GetUserJournal(ctx, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (j *JournalHandler) AddMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var mark entities.Mark
	if err := ctx.BindJSON(&mark); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	lesson, err := j.controller.AddMark(ctx, mark, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (j *JournalHandler) GetMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	group := ctx.Query("group")
	subject := ctx.Query("subject")
	userIdHex := ctx.Query("userId")

	lessons, err := j.controller.GetMark(ctx, group, subject, userIdHex, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lessons)
}

func (j *JournalHandler) UpdateMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var mark entities.Mark
	if err := ctx.BindJSON(&mark); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	lesson, err := j.controller.UpdateMark(ctx, mark, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (j *JournalHandler) DeleteMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	markId := ctx.Query("markId")
	userId := ctx.Query("userId")
	subjectId := ctx.Query("subjectId")

	lesson, err := j.controller.DeleteMark(ctx, markId, userId, subjectId, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}
